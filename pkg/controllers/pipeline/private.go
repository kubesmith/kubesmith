package pipeline

import (
	"context"
	"fmt"
	"time"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline/minio"
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/kubesmith/kubesmith/pkg/templates"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *PipelineController) processPipeline(action sync.SyncAction) error {
	cachedPipeline := action.GetObject().(api.Pipeline)
	logger := c.logger.WithFields(logrus.Fields{
		"Name": cachedPipeline.GetName(),
	})

	switch action.GetAction() {
	case sync.SyncActionDelete:
		return c.processDeletedPipeline(*cachedPipeline.DeepCopy(), logger)
	default:
		pipeline, err := c.pipelineLister.Pipelines(cachedPipeline.GetNamespace()).Get(cachedPipeline.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline")
		}

		// create a new logger for this pipeline's execution
		logger = logger.WithFields(logrus.Fields{
			"Phase":      pipeline.Status.Phase,
			"StageIndex": pipeline.Status.StageIndex,
		})

		// determine the phase and begin execution of the pipeline
		if pipeline.HasNoPhase() {
			return c.processEmptyPhasePipeline(*pipeline.DeepCopy(), logger)
		} else if pipeline.IsQueued() {
			return c.processQueuedPipeline(*pipeline.DeepCopy(), logger)
		} else if pipeline.IsRunning() {
			return c.processRunningPipeline(*pipeline.DeepCopy(), logger)
		} else if pipeline.HasSucceeded() {
			return c.processSuccessfulPipeline(*pipeline.DeepCopy(), logger)
		} else if pipeline.HasFailed() {
			return c.processFailedPipeline(*pipeline.DeepCopy(), logger)
		}
	}

	return nil
}

func (c *PipelineController) processEmptyPhasePipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *original.DeepCopy()

	logger.Info("validating")
	if err := pipeline.Validate(); err != nil {
		logger.Info("validation failed; marking as failed")

		pipeline.SetPhaseToFailed(err.Error())
		if _, err := c.patchPipeline(pipeline, original); err != nil {
			return errors.Wrap(err, "could not mark as failed")
		}

		logger.Info("marked as failed")
		return errors.Wrap(err, "validation failed")
	}

	logger.Info("validated; marking as queued")
	pipeline.SetPhaseToQueued()
	if _, err := c.patchPipeline(pipeline, original); err != nil {
		return errors.Wrap(err, "could not mark as queued")
	}

	logger.Info("marked as queued")
	return nil
}

func (c *PipelineController) processQueuedPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *original.DeepCopy()

	logger.Info("checking if another pipeline can be run")
	canRunAnotherPipeline, err := c.canRunAnotherPipeline(pipeline)
	if err != nil {
		return errors.Wrap(err, "could not check if another pipeline could be run")
	}

	if !canRunAnotherPipeline {
		logger.Info("cannot run another pipeline")
		return nil
	}

	logger.Info("marking as running")
	pipeline.SetPhaseToRunning()
	if _, err := c.patchPipeline(pipeline, original); err != nil {
		return errors.Wrap(err, "could not mark as running")
	}

	logger.Info("marked as running")
	return nil
}

func (c *PipelineController) processRunningPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	if original.Status.StageIndex > len(original.Spec.Jobs) {
		logger.Info("marking pipeline as succceeded")
		pipeline := *original.DeepCopy()
		pipeline.SetPhaseToSucceeded()

		if _, err := c.patchPipeline(pipeline, original); err != nil {
			return errors.Wrap(err, "could not mark pipeline as succeeded")
		}

		logger.Info("marked pipeline as succeeded")
		return nil
	}

	if err := c.ensureServiceAccountExists(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure service account exists")
	}

	if c.pipelineNeedsMinio(original) == true {
		minioServer, err := c.ensureMinioServerIsRunning(original, logger)
		if err != nil {
			return errors.Wrap(err, "could not ensure minio server is running")
		}

		logger.Info("updating pipeline with minio storage configuration")
		pipeline := c.setMinioInfoForPipeline(*original.DeepCopy(), minioServer)

		updated, err := c.patchPipeline(pipeline, original)
		if err != nil {
			return errors.Wrap(err, "could not update pipeline with minio storage configuration")
		}

		original = *updated.DeepCopy()
		logger.Info("updated pipeline with minio storage configuration")
	}

	if err := c.ensureRepoArtifactExists(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure repo artifact exists")
	}

	if err := c.ensureCurrentPipelineStageIsScheduled(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure current pipeline stage was scheduled")
	}

	return nil
}

func (c *PipelineController) setMinioInfoForPipeline(original api.Pipeline, minioServer *minio.MinioServer) api.Pipeline {
	host, _ := minioServer.GetServiceHost()
	original.Spec.Workspace.Storage.S3.Host = host
	original.Spec.Workspace.Storage.S3.Port = minio.MINIO_DEFAULT_PORT
	original.Spec.Workspace.Storage.S3.UseSSL = false
	original.Spec.Workspace.Storage.S3.BucketName = minioServer.GetBucketName()
	original.Spec.Workspace.Storage.S3.Credentials.Secret.Name = minioServer.GetResourceName()
	original.Spec.Workspace.Storage.S3.Credentials.Secret.AccessKeyKey = minio.MINIO_DEFAULT_ACCESS_KEY_KEY
	original.Spec.Workspace.Storage.S3.Credentials.Secret.SecretKeyKey = minio.MINIO_DEFAULT_SECRET_KEY_KEY

	return original
}

func (c *PipelineController) processSuccessfulPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("todo: processing successful pipeline")
	return nil
}

func (c *PipelineController) processFailedPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("todo: processing failed pipeline")
	return nil
}

func (c *PipelineController) processDeletedPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("cleaning up pipeline")
	if err := c.cleanupMinioServerForPipeline(original, logger); err != nil {
		return err
	}

	// create a selector for listing resources associated to pipelines
	labelSelector := c.getResourceLabelSelector(c.getWrappedLabels(original))

	// create the delete options that can help clean everything up
	propagationPolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	if err := c.deleteAssociatedPipelineStages(original, labelSelector, deleteOptions, logger); err != nil {
		return err
	}

	if err := c.deleteAssociatedJobs(original, labelSelector, deleteOptions, logger); err != nil {
		return err
	}

	if err := c.deleteAssociatedServiceAccount(original, deleteOptions, logger); err != nil {
		return err
	}

	logger.Info("pipeline cleaned up")
	return nil
}

func (c *PipelineController) deleteAssociatedJobs(
	original api.Pipeline,
	labelSelector labels.Selector,
	deleteOptions metav1.DeleteOptions,
	logger logrus.FieldLogger,
) error {
	logger.Info("retrieving jobs")
	jobs, err := c.jobLister.Jobs(original.GetNamespace()).List(labelSelector)
	if err != nil {
		return errors.Wrap(err, "could not retrieve jobs")
	}

	logger.Info("retrieved jobs")

	// delete all of the jobs associated to the pipeline
	logger.Info("deleting jobs")
	for _, job := range jobs {
		logger.WithFields(logrus.Fields{
			"JobName":      job.GetName(),
			"JobNamespace": job.GetNamespace(),
		}).Info("deleting job")

		if err := c.kubeClient.BatchV1().Jobs(job.GetNamespace()).Delete(job.GetName(), &deleteOptions); err != nil {
			return errors.Wrapf(err, "could not delete job: %s/%s", job.GetName(), job.GetNamespace())
		}
	}

	logger.Info("deleted jobs")
	return nil
}

func (c *PipelineController) deleteAssociatedServiceAccount(
	original api.Pipeline,
	deleteOptions metav1.DeleteOptions,
	logger logrus.FieldLogger,
) error {
	logger.Info("deleting service account (if it exists)")
	if err := c.kubeClient.CoreV1().ServiceAccounts(original.GetNamespace()).Delete(original.GetResourcePrefix(), &deleteOptions); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "could not delete service account")
		}
	}

	logger.Info("deleted service account")
	return nil
}

func (c *PipelineController) deleteAssociatedPipelineStages(
	original api.Pipeline,
	labelSelector labels.Selector,
	deleteOptions metav1.DeleteOptions,
	logger logrus.FieldLogger,
) error {
	logger.Info("retrieving pipeline stages")
	stages, err := c.pipelineStageLister.PipelineStages(original.GetNamespace()).List(labelSelector)
	if err != nil {
		return errors.Wrap(err, "could not retrieve pipeline stages")
	}

	logger.Info("retrieved pipeline stages")

	// delete all of the pipeline stages associated to the pipeline
	logger.Info("deleting pipeline stages")
	for _, stage := range stages {
		logger.WithFields(logrus.Fields{
			"PipelineStageName":      stage.GetName(),
			"PipelineStageNamespace": stage.GetNamespace(),
		}).Info("deleting pipeline stage")

		if err := c.kubesmithClient.PipelineStages(stage.GetNamespace()).Delete(stage.GetName(), &deleteOptions); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}

			return errors.Wrapf(err, "could not delete pipeline stage: %s/%s", stage.GetName(), stage.GetNamespace())
		}
	}

	logger.Info("deleted pipeline stages")
	return nil
}

func (c *PipelineController) getResourceLabelSelector(resourceLabels map[string]string) labels.Selector {
	set := labels.Set{}

	for key, value := range resourceLabels {
		set[key] = value
	}

	return labels.SelectorFromSet(set)
}

func (c *PipelineController) patchPipeline(updated, original api.Pipeline) (*api.Pipeline, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.Pipelines(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}

func (c *PipelineController) canRunAnotherPipeline(original api.Pipeline) (bool, error) {
	pipelines, err := c.pipelineLister.Pipelines(original.GetNamespace()).List(labels.Everything())
	if err != nil {
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines {
		if pipeline.IsRunning() {
			currentlyRunning++
		}
	}

	if currentlyRunning < c.maxRunningPipelines {
		return true, nil
	}

	return false, nil
}

func (c *PipelineController) getWrappedLabels(pipeline api.Pipeline) map[string]string {
	labels := pipeline.GetResourceLabels()
	labels[api.GetLabelKey("Controller")] = "Pipeline"

	return labels
}

func (c *PipelineController) ensureServiceAccountExists(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("ensuring service account exists")
	if _, err := c.serviceAccountLister.ServiceAccounts(original.GetNamespace()).Get(original.GetResourcePrefix()); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("service account does not exist; scheduling...")

			serviceAccount := templates.GetPipelineServiceAccount(original)
			if _, err := c.kubeClient.CoreV1().ServiceAccounts(original.GetNamespace()).Create(&serviceAccount); err != nil {
				return errors.Wrap(err, "could not schedule service account")
			}

			logger.Info("service account is now scheduled")
			return nil
		}

		return errors.Wrap(err, "could not get service account")
	}

	logger.Info("service account is scheduled")
	return nil
}

func (c *PipelineController) ensureMinioServerIsRunning(original api.Pipeline, logger logrus.FieldLogger) (*minio.MinioServer, error) {
	logger.Info("ensuring minio is scheduled")
	minioServer := minio.NewMinioServer(
		original.GetNamespace(),
		original.GetResourcePrefix(),
		logger,
		c.getWrappedLabels(original),
		c.kubeClient,
		c.secretLister,
		c.deploymentLister,
		c.serviceLister,
	)

	if err := minioServer.Create(); err != nil {
		return nil, errors.Wrap(err, "could not create minio server")
	}

	logger.Info("minio scheduled")
	logger.Info("waiting for minio availability")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	minioServerAvailable := make(chan bool, 1)
	go minioServer.WaitForAvailability(ctx, 5, minioServerAvailable)

	if available := <-minioServerAvailable; !available {
		return nil, errors.New("minio server is not available")
	}

	logger.Info("minio server is available")
	return minioServer, nil
}

func (c *PipelineController) ensureRepoArtifactExists(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("fetching secret data")
	secret, err := c.secretLister.Secrets(original.GetNamespace()).Get(original.Spec.Workspace.Storage.S3.Credentials.Secret.Name)
	if err != nil {
		return errors.Wrap(err, "could not fetch secret information")
	}
	logger.Info("fetched secret data")

	s3Client, err := s3.NewS3Client(
		// todo: original.Spec.Workspace.Storage.S3.Host,
		"127.0.0.1",
		original.Spec.Workspace.Storage.S3.Port,
		string(secret.Data[original.Spec.Workspace.Storage.S3.Credentials.Secret.AccessKeyKey]),
		string(secret.Data[original.Spec.Workspace.Storage.S3.Credentials.Secret.SecretKeyKey]),
		false,
	)

	if err != nil {
		return errors.Wrap(err, "could not create an s3 client")
	}

	logger.Info("checking for repo artifact")
	fileName := fmt.Sprintf("%s/repo/repo.tar.gz", original.GetResourcePrefix())
	exists, err := s3Client.FileExists(original.Spec.Workspace.Storage.S3.BucketName, fileName)
	if err != nil {
		return errors.Wrap(err, "could not check for repo artifact")
	} else if exists {
		logger.Info("repo artifact exists")
		return nil
	}

	logger.Info("repo artifact does not exist")
	if err := c.ensureRepoArtifactJobIsScheduled(original, logger); err != nil {
		return err
	}

	logger.Info("waiting for repo artifact to be created")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	repoArtifactCreated := make(chan bool, 1)
	go c.waitForRepoArtifactToBeCreated(
		ctx,
		original.GetResourcePrefix(),
		2,
		original.Spec.Workspace.Storage.S3.BucketName,
		s3Client,
		repoArtifactCreated,
	)

	if exists := <-repoArtifactCreated; !exists {
		logger.Info("repo artifact was not created; requeueing pipeline")
		new := *original.DeepCopy()
		if _, err := c.patchPipeline(new, original); err != nil {
			return errors.Wrap(err, "could not requeue pipeline")
		}

		logger.Info("pipeline requeued")
		return err
	}

	logger.Info("repo artifact was created")
	return nil
}

func (c *PipelineController) ensureRepoArtifactJobIsScheduled(original api.Pipeline, logger logrus.FieldLogger) error {
	pipelineName := original.GetResourcePrefix()
	name := fmt.Sprintf("%s-clone-repo", pipelineName)

	// check to see if the job already exists
	logger.Info("ensuring clone repo job is scheduled")
	if _, err := c.jobLister.Jobs(original.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("clone repo job was not found; scheduling...")

			original.ObjectMeta.Labels = c.getWrappedLabels(original)
			job := templates.GetJobCloneRepo(original)
			if _, err := c.kubeClient.BatchV1().Jobs(original.GetNamespace()).Create(&job); err != nil {
				return errors.Wrap(err, "could not schedule clone repo job")
			}

			logger.Info("clone repo job is now scheduled")
			return nil
		}

		return errors.Wrap(err, "could not fetch clone repo job")
	}

	logger.Info("clone repo job is scheduled")
	return nil
}

func (c *PipelineController) waitForRepoArtifactToBeCreated(
	ctx context.Context,
	remotePath string,
	secondsInterval int,
	bucketName string,
	s3Client *s3.S3Client,
	repoArtifactCreated chan bool,
) {
	artifactRemotePath := fmt.Sprintf("%s/repo/repo.tar.gz", remotePath)

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				repoArtifactCreated <- false
			}
		default:
			if exists, _ := s3Client.FileExists(bucketName, artifactRemotePath); exists {
				repoArtifactCreated <- true
				return
			}
		}

		time.Sleep(time.Second * time.Duration(secondsInterval))
	}
}

func (c *PipelineController) cleanupMinioServerForPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("cleaning up minio server resources")
	minioServer := minio.NewMinioServer(
		original.GetNamespace(),
		original.GetResourcePrefix(),
		logger,
		c.getWrappedLabels(original),
		c.kubeClient,
		c.secretLister,
		c.deploymentLister,
		c.serviceLister,
	)

	if err := minioServer.Delete(); err != nil {
		return errors.Wrap(err, "could not cleanup minio server resources")
	}

	logger.Info("cleaned up minio server resources")
	return nil
}

func (c *PipelineController) ensureCurrentPipelineStageIsScheduled(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("ensuring pipeline stage is scheduled")
	name := fmt.Sprintf("%s-stage-%d", original.GetResourcePrefix(), original.Status.StageIndex)

	if _, err := c.pipelineStageLister.PipelineStages(original.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("pipeline stage was not found; scheduling...")

			pipelineStage := GetPipelineStage(
				name,
				original.GetWorkspacePath(),
				original.Spec.Workspace.Storage,
				c.getWrappedLabels(original),
				original.GetExpandedJobsForCurrentStage(),
			)

			if _, err := c.kubesmithClient.PipelineStages(original.GetNamespace()).Create(&pipelineStage); err != nil {
				return errors.Wrap(err, "could not schedule pipeline stage")
			}

			logger.Info("pipeline stage is now scheduled")
			return nil
		}

		return errors.Wrap(err, "could not get pipeline stage")
	}

	logger.Info("pipeline stage is scheduled")
	return nil
}

func (c *PipelineController) pipelineNeedsMinio(original api.Pipeline) bool {
	s3 := original.Spec.Workspace.Storage.S3

	return (s3.Host == "") ||
		(s3.Port <= 0) ||
		(s3.BucketName == "") ||
		(s3.Credentials.Secret.Name == "") ||
		(s3.Credentials.Secret.AccessKeyKey == "") ||
		(s3.Credentials.Secret.SecretKeyKey == "")
}
