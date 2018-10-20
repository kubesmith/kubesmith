package pipeline

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	logger.Info("todo: processing queued pipeline")
	return nil
}

func (c *PipelineController) processRunningPipeline(original api.Pipeline, logger logrus.FieldLogger) error {
	logger.Info("todo: processing running pipeline")
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
	logger.Info("todo: processing deleted pipeline")
	return nil
}

func (c *PipelineController) patchPipeline(updated, original api.Pipeline) (*api.Pipeline, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.Pipelines(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
