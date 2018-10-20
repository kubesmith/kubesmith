package pipeline

import (
	"context"
	"time"

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
		c.logger.Panic(errors.New("programmer error; pipeline is nil"))
	}

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": pipeline.GetName(),
		})

		if err := c.processDeletedPipeline(*pipeline.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		pipeline, err := c.pipelineLister.Pipelines(pipeline.GetNamespace()).Get(pipeline.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline")
		}

		// create a new logger for this pipeline's execution
		logger := c.logger.WithFields(logrus.Fields{
			"Name":       pipeline.GetName(),
			"Phase":      pipeline.GetPhase(),
			"StageIndex": pipeline.GetStageIndex(),
		})

		// determine the phase and begin execution of the pipeline
		if pipeline.HasNoPhase() {
			if err := c.processEmptyPhasePipeline(*pipeline.DeepCopy(), logger); err != nil {
				return err
			}
		} else if pipeline.IsQueued() {
			if err := c.processQueuedPipeline(*pipeline.DeepCopy(), logger); err != nil {
				return err
			}
		} else if pipeline.IsRunning() {
			if err := c.processRunningPipeline(*pipeline.DeepCopy(), logger); err != nil {
				return err
			}
		} else if pipeline.HasSucceeded() {
			if err := c.processSuccessfulPipeline(*pipeline.DeepCopy(), logger); err != nil {
				return err
			}
		} else if pipeline.HasFailed() {
			if err := c.processFailedPipeline(*pipeline.DeepCopy(), logger); err != nil {
				return err
			}
		}

		break
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
		logger.Warn("cannot run another pipeline")
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
	minioServer, err := c.ensureMinioServerIsRunning(original, logger)
	if err != nil {
		return errors.Wrap(err, "could not ensure minio server is running")
	}

	if err := c.ensureRepoArtifactExists(original, minioServer, logger); err != nil {
		return errors.Wrap(err, "could not ensure repo artifact exists")
	}

	return nil
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

	// create a selector for listing resources associated to pipeline jobs
	logger.Info("creating label selector for associated resources")
	labelSelector, err := original.GetResourceLabelSelector()
	if err != nil {
		return errors.Wrap(err, "could not create label selector for pipeline")
	}
	logger.Info("created label selector for associated resources")

	// create the delete options that can help clean everything up
	propagationPolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	// attempt to get all pipeline stages associated to the pipeline
	logger.Info("retrieving pipeline stages")
	stages, err := c.pipelineStageLister.PipelineStages(original.GetNamespace()).List(labelSelector)
	if err != nil {
		return errors.Wrap(err, "could not retrieve pipeline stages")
	}
	logger.Info("retrieved pipeline stages")

	// delete all of the pipeline stages associated to the pipeline
	logger.Info("deleting pipeline stages")
	for _, stage := range stages {
		if err := c.kubesmithClient.PipelineStages(stage.GetNamespace()).Delete(stage.GetName(), &deleteOptions); err != nil {
			return errors.Wrapf(err, "could not delete pipeline stage: %s/%s", stage.GetName(), stage.GetNamespace())
		}
	}

	logger.Info("deleted pipeline stages")
	logger.Info("pipeline cleaned up")
	return nil
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

func (c *PipelineController) ensureMinioServerIsRunning(original api.Pipeline, logger logrus.FieldLogger) (*minio.MinioServer, error) {
	logger.Info("scheduling minio...")
	minioServer := minio.NewMinioServer(
		original.GetNamespace(),
		original.GetResourcePrefix(),
		logger,
		original.GetResourceLabels(),
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

func (c *PipelineController) ensureRepoArtifactExists(original api.Pipeline, minioServer *minio.MinioServer, logger logrus.FieldLogger) error {
	return nil
}

func (c *PipelineController) cleanupMinioServerForPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("cleaning up minio server resources")
	minioServer := minio.NewMinioServer(
		original.GetNamespace(),
		original.GetResourcePrefix(),
		logger,
		original.GetResourceLabels(),
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
