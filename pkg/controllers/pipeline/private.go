package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kubesmith/kubesmith/pkg/s3"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/annotations"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline/minio"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *PipelineController) processPipeline(action sync.SyncAction) error {
	pipeline := action.GetObject().(*api.Pipeline)
	if pipeline == nil {
		c.logger.Panic(errors.New("programmer error; pipeline object was nil"))
	}

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": pipeline.Name,
		})

		if err := c.processDeletedPipeline(*pipeline, logger); err != nil {
			return err
		}
	default:
		pipeline, err := c.pipelineLister.Pipelines(pipeline.GetNamespace()).Get(pipeline.GetName())
		if apierrors.IsNotFound(err) {
			glog.V(1).Info("unable to find pipeline")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline")
		}

		// create a new logger for this pipeline's execution
		fieldLogger := c.logger.WithFields(logrus.Fields{
			"Name":       pipeline.Name,
			"Phase":      pipeline.Status.Phase,
			"StageIndex": pipeline.Status.StageIndex,
		})

		// determine the phase and begin execution of the pipeline
		if pipeline.HasNoPhase() {
			if err := c.processEmptyPhasePipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
				return err
			}
		} else if pipeline.IsQueued() {
			if err := c.processQueuedPipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
				return err
			}
		} else if pipeline.IsRunning() {
			if err := c.processRunningPipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
				return err
			}
		} else if pipeline.HasCompleted() {
			if err := c.processCompletedPipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
				return err
			}
		} else if pipeline.HasFailed() {
			if err := c.processFailedPipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *PipelineController) resyncPipelines() {
	pipelines, err := c.pipelineLister.List(labels.Everything())
	if err != nil {
		glog.Error(errors.Wrap(err, "error listing pipelines"))
		return
	}

	for _, pipeline := range pipelines {
		c.Queue.Add(sync.PipelineUpdateAction(pipeline))
	}
}

func (c *PipelineController) processEmptyPhasePipeline(originalPipeline api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *originalPipeline.DeepCopy()

	logger.Info("validating")
	if err := pipeline.Validate(); err != nil {
		err = errors.Wrap(err, "could not validate")
		logger.Info("marking as failed")

		pipeline.SetPhaseToFailed(err.Error())
		if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
			logger.Error(errors.Wrap(err, "could not mark as failed"))
		}

		logger.Info("marked as failed!")
		return err
	}

	logger.Info("validated; marking as queued")
	pipeline.SetPhaseToQueued()
	if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
		return errors.Wrap(err, "could not mark as queued")
	}

	logger.Info("marked as queued")
	return nil
}

func (c *PipelineController) processQueuedPipeline(originalPipeline api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *originalPipeline.DeepCopy()

	logger.Info("checking if another pipeline can be run")
	canRunAnotherPipeline, err := c.canRunAnotherPipeline(pipeline)
	if err != nil {
		logger.Error(errors.Wrap(err, "could not check if another pipeline could be run"))
		return err
	}

	if !canRunAnotherPipeline {
		logger.Warn("cannot run another pipeline")
		return nil
	}

	logger.Info("marking as running")
	pipeline.SetPhaseToRunning()
	if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
		logger.Error(errors.Wrap(err, "could not mark as running"))
		return err
	}

	logger.Info("marked as running")
	return nil
}

func (c *PipelineController) processRunningPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	minioServer, err := c.ensureMinioServerIsRunning(pipeline, logger)
	if err != nil {
		return errors.Wrap(err, "could not ensure minio server is running")
	}

	if err := c.ensureRepoArtifactExists(pipeline, minioServer, logger); err != nil {
		return errors.Wrap(err, "could not ensure repo artifact exists")
	}

	logger.Info("ensuring jobs are scheduled")
	for index, job := range pipeline.GetExpandedJobsForCurrentStage() {
		jobIndex := index + 1
		tmpLogger := logger.WithField("JobIndex", jobIndex)

		if err := c.ensureJobIsScheduled(pipeline, job, jobIndex, minioServer, tmpLogger); err != nil {
			return err
		}
	}

	logger.Info("jobs are scheduled")
	return nil
}

func (c *PipelineController) processCompletedPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	logger.Warn("todo: processing completed pipeline")
	return nil
}

func (c *PipelineController) processFailedPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	logger.Warn("todo: processing failed pipeline")
	return nil
}

func (c *PipelineController) processDeletedPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	if err := c.cleanupMinioServerForPipeline(pipeline, logger); err != nil {
		return err
	}

	// create a selector for listing resources associated to pipeline jobs
	logger.Info("creating label selector for associated resources")
	labelSelector, err := pipeline.GetResourceLabelSelector()
	if err != nil {
		return errors.Wrap(err, "could not create label selector for pipeline")
	}
	logger.Info("created label selector for associated resources")

	// create the delete options that can help clean everything up
	propagationPolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	// attempt to get all Jobs associated to the pipeline
	logger.Info("retrieving jobs")
	jobs, err := c.jobLister.Jobs(pipeline.GetNamespace()).List(labelSelector)
	if err != nil {
		err = errors.Wrap(err, "could not retrieve jobs")
		logger.Error(err)
		return err
	}
	logger.Info("retrieved jobs")

	// delete all of the jobs associated to the pipeline
	logger.Info("deleting jobs")
	for _, job := range jobs {
		if err := c.kubeClient.BatchV1().Jobs(job.GetNamespace()).Delete(job.GetName(), &deleteOptions); err != nil {
			err = errors.Wrap(err, "could not delete job")
			logger.WithFields(logrus.Fields{
				"JobName":      job.GetName(),
				"JobNamespace": job.GetNamespace(),
			}).Error(err)
			return err
		}
	}
	logger.Info("deleted jobs")

	// attempt to get all ConfigMaps associated to the pipeline
	logger.Info("retrieving configmaps")
	configMaps, err := c.configMapLister.ConfigMaps(pipeline.GetNamespace()).List(labelSelector)
	if err != nil {
		err = errors.Wrap(err, "could not retrieve configmaps")
		logger.Error(err)
		return err
	}
	logger.Info("retrieved configmaps")

	// delete all of the configmaps associated to the pipeline
	logger.Info("deleting configmaps")
	for _, configMap := range configMaps {
		if err := c.kubeClient.CoreV1().ConfigMaps(configMap.GetNamespace()).Delete(configMap.GetName(), &deleteOptions); err != nil {
			err = errors.Wrap(err, "could not delete configmap")
			logger.WithFields(logrus.Fields{
				"ConfigMapName":      configMap.GetName(),
				"ConfigMapNamespace": configMap.GetNamespace(),
			}).Error(err)
			return err
		}
	}

	logger.Info("deleted configmaps")
	logger.Info("pipeline cleaned up")
	return nil
}

func (c *PipelineController) cleanupMinioServerForPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("cleaning up minio server resources")
	minioServer := minio.NewMinioServer(
		pipeline.GetNamespace(),
		pipeline.GetResourcePrefix(),
		pipeline.GetResourceLabels(),
		logger,
		c.kubeClient,
		c.deploymentLister,
	)

	if err := minioServer.Delete(); err != nil {
		return errors.Wrap(err, "could not cleanup minio server resources")
	}

	logger.Info("cleaned up minio server resources")
	return nil
}

func (c *PipelineController) canRunAnotherPipeline(pipeline api.Pipeline) (bool, error) {
	pipelines, err := c.pipelineLister.Pipelines(pipeline.GetNamespace()).List(labels.Everything())
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

func (c *PipelineController) patchPipeline(updated, original api.Pipeline) (*api.Pipeline, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.Pipelines(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}

func (c *PipelineController) getJobResourceName(pipeline api.Pipeline, jobIndex int) string {
	return fmt.Sprintf(
		"%s-stage-%d-job-%d",
		pipeline.GetResourcePrefix(),
		pipeline.GetStageIndex(),
		jobIndex,
	)
}

func (c *PipelineController) ensureMinioServerIsRunning(pipeline api.Pipeline, logger logrus.FieldLogger) (*minio.MinioServer, error) {
	logger.Info("scheduling minio...")
	minioServer := minio.NewMinioServer(
		pipeline.GetNamespace(),
		pipeline.GetResourcePrefix(),
		pipeline.GetResourceLabels(),
		logger,
		c.kubeClient,
		c.deploymentLister,
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

func (c *PipelineController) ensureRepoArtifactExists(pipeline api.Pipeline, minioServer *minio.MinioServer, logger logrus.FieldLogger) error {
	s3Client, err := minioServer.GetS3Client()
	if err != nil {
		return errors.Wrap(err, "could not get s3 client")
	}

	logger.Info("checking for repo artifact")
	exists, err := s3Client.FileExists("workspace", "repo.tar.gz")
	if err != nil {
		logger.Error(errors.Wrap(err, "could not check for repo artifact"))
		return err
	} else if exists {
		logger.Info("repo artifact exists")
		return nil
	}

	logger.Info("repo artifact does not exist")
	if err := c.ensureRepoArtifactJobIsScheduled(pipeline, minioServer, logger); err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("waiting for repo artifact to be created")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	repoArtifactCreated := make(chan bool, 1)
	go c.waitForRepoArtifactToBeCreated(ctx, 5, s3Client, repoArtifactCreated)

	if exists := <-repoArtifactCreated; !exists {
		logger.Info("repo artifact was not created; requeueing pipeline")
		new := *pipeline.DeepCopy()
		if _, err := c.patchPipeline(new, pipeline); err != nil {
			logger.Error(errors.Wrap(err, "could not requeue pipeline"))
			return err
		}

		logger.Info("pipeline requeued")
		return err
	}

	logger.Info("repo artifact was created")
	return nil
}

func (c *PipelineController) ensureRepoArtifactJobIsScheduled(pipeline api.Pipeline, minioServer *minio.MinioServer, logger logrus.FieldLogger) error {
	name := fmt.Sprintf("%s-clone-repo", pipeline.GetResourcePrefix())

	// check to see if the job already exists
	logger.Info("ensuring clone repo job is scheduled")
	if _, err := c.jobLister.Jobs(pipeline.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("clone repo job was not found; scheduling...")

			host, err := minioServer.GetServiceHost()
			if err != nil {
				return err
			}

			job := GetRepoCloneJob(
				name,
				pipeline.GetWorkspaceSSHSecretName(),
				pipeline.GetWorkspaceSSHSecretKey(),
				pipeline.GetWorkspaceRepoURL(),
				host,
				9000,
				minioServer.GetResourceName(),
				pipeline.GetResourceLabels(),
			)

			if _, err := c.kubeClient.BatchV1().Jobs(pipeline.GetNamespace()).Create(&job); err != nil {
				return errors.Wrap(err, "could not create clone repo job")
			}

			logger.Info("clone repo job is now scheduled")
			return nil
		}

		return errors.Wrap(err, "could not fetch clone repo job")
	}

	logger.Info("clone repo job is scheduled")
	return nil
}

func (c *PipelineController) waitForRepoArtifactToBeCreated(ctx context.Context, secondsInterval int, s3Client *s3.S3Client, repoArtifactCreated chan bool) {
	for {
		select {
		case <-ctx.Done():
			repoArtifactCreated <- false
		default:
			if exists, _ := s3Client.FileExists("workspace", "repo.tar.gz"); exists {
				repoArtifactCreated <- true
			}
		}

		time.Sleep(time.Second * time.Duration(secondsInterval))
	}
}

func (c *PipelineController) getPipelineJobConfigMapData(job api.PipelineSpecJob) map[string]string {
	if len(job.ConfigMapData) > 0 {
		return job.ConfigMapData
	}

	return map[string]string{"pipeline-script.sh": strings.Join(job.Runner, "\n")}
}

func (c *PipelineController) getPipelineJobCommand(job api.PipelineSpecJob) []string {
	if len(job.Command) > 0 {
		return job.Command
	}

	return []string{"/bin/sh", "-x", "/kubesmith/scripts/pipeline-script.sh"}
}

func (c *PipelineController) getPipelineJobArgs(job api.PipelineSpecJob) []string {
	if len(job.Args) > 0 {
		return job.Args
	}

	return []string{}
}

func (c *PipelineController) ensureJobConfigMapExists(
	pipeline api.Pipeline,
	name string,
	configMapData map[string]string,
	logger logrus.FieldLogger,
) error {
	logger.Info("checking for job configmap")
	if _, err := c.configMapLister.ConfigMaps(pipeline.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("job configmap does not exist; creating")
			resource := GetJobConfigMap(name, pipeline.GetResourceLabels(), configMapData)
			_, err := c.kubeClient.CoreV1().ConfigMaps(pipeline.GetNamespace()).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create job configmap")
			}

			logger.Info("job configmap created")
			return nil
		}

		return errors.Wrap(err, "could not get job configmap")
	}

	logger.Info("job configmap exists")
	return nil
}

func (c *PipelineController) ensureJobIsScheduled(
	pipeline api.Pipeline,
	job api.PipelineSpecJob,
	jobIndex int,
	minioServer *minio.MinioServer,
	logger logrus.FieldLogger,
) error {
	name := c.getJobResourceName(pipeline, jobIndex)
	configMapData := c.getPipelineJobConfigMapData(job)

	// ensure the configMap exists (if it's needed)
	if len(configMapData) > 0 {
		if err := c.ensureJobConfigMapExists(pipeline, name, configMapData, logger); err != nil {
			return err
		}
	}

	// now, ensure the job exists
	logger.Info("checking to see if job exists...")
	if _, err := c.jobLister.Jobs(pipeline.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			jobAnnotations := map[string]string{}

			if job.AllowFailure == true {
				jobAnnotations[annotations.AllowFailure] = "true"
			}

			s3Host, err := minioServer.GetServiceHost()
			if err != nil {
				return err
			}

			resource := GetJob(
				name,
				job.Image,
				s3Host,
				9000,
				pipeline.GetResourcePrefix(),
				minioServer.GetResourceName(),
				pipeline.GetWorkspacePath(),
				jobAnnotations,
				job.Environment,
				c.getPipelineJobCommand(job),
				c.getPipelineJobArgs(job),
				pipeline.GetResourceLabels(),
			)

			if _, err := c.kubeClient.BatchV1().Jobs(pipeline.GetNamespace()).Create(&resource); err != nil {
				return errors.Wrap(err, "could not create job")
			}

			logger.Info("job created")
			return nil
		}

		return errors.Wrap(err, "could not find job")
	}

	logger.Info("job exists")
	return nil
}
