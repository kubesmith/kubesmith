package pipelinestage

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *PipelineStageController) processPipelineStage(action sync.SyncAction) error {
	stage := action.GetObject().(*api.PipelineStage)
	if stage == nil {
		c.logger.Panic(errors.New("programmer error; pipeline stage is nil"))
	}

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": stage.GetName(),
		})

		if err := c.processDeletedPipelineStage(*stage.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		stage, err := c.pipelineStageLister.PipelineStages(stage.GetNamespace()).Get(stage.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline stage")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline stage")
		}

		// create a new logger for this pipeline stage's execution
		logger := c.logger.WithFields(logrus.Fields{
			"Name":  stage.GetName(),
			"Phase": stage.GetPhase(),
		})

		// determine the phase and begin execution of the pipeline stage
		if stage.HasNoPhase() {
			if err := c.processEmptyPhasePipelineStage(*stage.DeepCopy(), logger); err != nil {
				return err
			}
		} else if stage.IsRunning() {
			if err := c.processRunningPipelineStage(*stage.DeepCopy(), logger); err != nil {
				return err
			}
		} else if stage.HasSucceeded() {
			if err := c.processSuccessfulPipelineStage(*stage.DeepCopy(), logger); err != nil {
				return err
			}
		} else if stage.HasFailed() {
			if err := c.processFailedPipelineStage(*stage.DeepCopy(), logger); err != nil {
				return err
			}
		}

		break
	}

	return nil
}

func (c *PipelineStageController) processEmptyPhasePipelineStage(original api.PipelineStage, logger logrus.FieldLogger) error {
	stage := *original.DeepCopy()

	logger.Info("validating")
	if err := stage.Validate(); err != nil {
		logger.Info("validation failed; marking as failed")

		stage.SetPhaseToFailed(err.Error())
		if _, err := c.patchPipelineStage(stage, original); err != nil {
			return errors.Wrap(err, "could not mark as failed")
		}

		logger.Info("marked as failed")
		return errors.Wrap(err, "validation failed")
	}

	logger.Info("validated; marking as running")
	stage.SetPhaseToRunning()
	if _, err := c.patchPipelineStage(stage, original); err != nil {
		return errors.Wrap(err, "could not mark as queued")
	}

	logger.Info("marked as running")
	return nil
}

func (c *PipelineStageController) processRunningPipelineStage(original api.PipelineStage, logger logrus.FieldLogger) error {
	logger.Info("todo: processing running pipeline stage")
	return nil
}

func (c *PipelineStageController) processSuccessfulPipelineStage(original api.PipelineStage, logger logrus.FieldLogger) error {
	logger.Info("todo: processing successful pipeline stage")
	return nil
}

func (c *PipelineStageController) processFailedPipelineStage(original api.PipelineStage, logger logrus.FieldLogger) error {
	logger.Info("todo: processing failed pipeline stage")
	return nil
}

func (c *PipelineStageController) processDeletedPipelineStage(original api.PipelineStage, logger logrus.FieldLogger) error {
	logger.Info("todo: processing deleted pipeline stage")
	return nil
}

func (c *PipelineStageController) patchPipelineStage(updated, original api.PipelineStage) (*api.PipelineStage, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineStages(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
