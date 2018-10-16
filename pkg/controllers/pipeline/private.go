package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline/minio"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func (c *PipelineController) processPipeline(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	pipeline, err := c.pipelineLister.Pipelines(ns).Get(name)
	if apierrors.IsNotFound(err) {
		glog.V(1).Info("unable to find pipeline")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting pipeline")
	}

	// create a new logger for this pipeline's execution
	fieldLogger := c.logger.WithFields(logrus.Fields{
		"Name":       pipeline.Name,
		"Namespace":  pipeline.Namespace,
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
	} else if pipeline.HasCompleted() || pipeline.HasFailed() {
		if err := c.processFinishedPipeline(*pipeline.DeepCopy(), fieldLogger); err != nil {
			return err
		}
	}

	return nil
}

func (c *PipelineController) resync() {
	list, err := c.kubesmithClient.Pipelines(c.namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.V(1).Info("error listing pipelines")
		glog.Error(err)
		return
	}

	for _, forge := range list.Items {
		key, err := cache.MetaNamespaceKeyFunc(forge)
		if err != nil {
			glog.Errorf("error generating key for pipeline; key: %s", forge.Name)
			continue
		}

		c.Queue.Add(key)
	}
}

func (c *PipelineController) processEmptyPhasePipeline(originalPipeline api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *originalPipeline.DeepCopy()

	logger.Info("validating pipeline...")
	if err := pipeline.Validate(); err != nil {
		logger.Error(errors.Wrap(err, "could not validate pipeline"))
		logger.Info("marking pipeline as failed")

		pipeline.SetPipelineToFailed(err.Error())
		if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
			logger.Error(errors.Wrap(err, "could not mark pipeline as failed"))
		}

		logger.Info("marked pipeline as failed!")
		return err
	}

	logger.Info("finished validating pipeline")
	logger.Info("marking pipeline as queued...")

	pipeline.SetPipelineToQueued()
	if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
		logger.Error(errors.Wrap(err, "could not set pipeline to running"))
		return err
	}

	logger.Info("marked pipeline as queued")
	return nil
}

func (c *PipelineController) processQueuedPipeline(originalPipeline api.Pipeline, logger logrus.FieldLogger) error {
	pipeline := *originalPipeline.DeepCopy()

	logger.Info("checking to see if we can run another pipeline in this namespace")
	canRunAnotherPipeline, err := c.canRunAnotherPipeline(pipeline)
	if err != nil {
		logger.Error(errors.Wrap(err, "could not check to see if we could run another pipeline"))
		return err
	}

	if !canRunAnotherPipeline {
		logger.Warn("cannot run another pipeline in this namespace")
		return nil
	}

	logger.Info("marking pipeline as running...")
	pipeline.SetPipelineToRunning()
	if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
		logger.Error(errors.Wrap(err, "could not set pipeline to running"))
		return err
	}

	logger.Info("marked pipeline as running")
	return nil
}

func (c *PipelineController) processRunningPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	minioServer, err := c.ensureMinioServerIsRunning(pipeline, logger)
	if err != nil {
		return err
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

func (c *PipelineController) processFinishedPipeline(pipeline api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("todo: processing finished pipeline")
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
	logger.Info("ensuring minio server resources are scheduled...")
	minioServer := minio.NewMinioServer(
		pipeline.GetNamespace(),
		pipeline.GetResourcePrefix(),
		pipeline.GetResourceLabels(),
		logger,
		c.kubeClient,
		c.deploymentLister,
	)

	if err := minioServer.Create(); err != nil {
		logger.Error(errors.Wrap(err, "could not ensure minio server resources are scheduled"))
		return nil, err
	}

	logger.Info("minio server resources are scheduled")
	logger.Info("waiting for minio server to be available...")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	minioServerAvailable := make(chan bool, 1)
	go minioServer.WaitForAvailability(ctx, 5, minioServerAvailable)

	if available := <-minioServerAvailable; !available {
		return nil, errors.New("minio server is not available")
	}

	logger.Info("minio server is available!")
	return minioServer, nil
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
	logger.Info("checking to see if configmap for job exists...")
	if _, err := c.configMapLister.ConfigMaps(pipeline.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			resource := GetJobConfigMap(name, pipeline.GetResourceLabels(), configMapData)
			_, err := c.kubeClient.CoreV1().ConfigMaps(pipeline.GetNamespace()).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create configmap for job")
			}

			logger.Info("configmap for job created")
			return nil
		}

		return errors.Wrap(err, "could not find configmap for job")
	}

	logger.Info("configmap for job exists")
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
			resource := GetJob(
				name,
				job.Image,
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
