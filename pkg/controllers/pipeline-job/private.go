package pipelinejob

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *PipelineJobController) processPipelineJob(action sync.SyncAction) error {
	job := action.GetObject().(*api.PipelineJob)
	if job == nil {
		c.logger.Panic(errors.New("programmer error; pipeline job is nil"))
	}

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": job.GetName(),
		})

		if err := c.processDeletedPipelineJob(*job.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		job, err := c.pipelineJobLister.PipelineJobs(job.GetNamespace()).Get(job.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline job")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline job")
		}

		// create a new logger for this pipeline job's execution
		logger := c.logger.WithFields(logrus.Fields{
			"Name":  job.GetName(),
			"Phase": job.GetPhase(),
		})

		// determine the phase and begin execution of the pipeline job
		if job.HasNoPhase() {
			if err := c.processEmptyPhasePipelineJob(*job.DeepCopy(), logger); err != nil {
				return err
			}
		} else if job.IsQueued() {
			if err := c.processQueuedPipelineJob(*job.DeepCopy(), logger); err != nil {
				return err
			}
		} else if job.IsRunning() {
			if err := c.processRunningPipelineJob(*job.DeepCopy(), logger); err != nil {
				return err
			}
		} else if job.HasSucceeded() {
			if err := c.processSuccessfulPipelineJob(*job.DeepCopy(), logger); err != nil {
				return err
			}
		} else if job.HasFailed() {
			if err := c.processFailedPipelineJob(*job.DeepCopy(), logger); err != nil {
				return err
			}
		}

		break
	}

	return nil
}

func (c *PipelineJobController) processEmptyPhasePipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	job := *original.DeepCopy()

	logger.Info("validating")
	if err := job.Validate(); err != nil {
		logger.Info("validation failed; marking as failed")

		job.SetPhaseToFailed(err.Error())
		if _, err := c.patchPipelineJob(job, original); err != nil {
			return errors.Wrap(err, "could not mark as failed")
		}

		logger.Info("marked as failed")
		return errors.Wrap(err, "validation failed")
	}

	logger.Info("validated; marking as queued")
	job.SetPhaseToQueued()
	if _, err := c.patchPipelineJob(job, original); err != nil {
		return errors.Wrap(err, "could not mark as queued")
	}

	logger.Info("marked as queued")
	return nil
}

func (c *PipelineJobController) processQueuedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing queued pipeline")
	return nil
}

func (c *PipelineJobController) processRunningPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing running pipeline")
	return nil
}

func (c *PipelineJobController) processSuccessfulPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing successful pipeline")
	return nil
}

func (c *PipelineJobController) processFailedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing failed pipeline")
	return nil
}

func (c *PipelineJobController) processDeletedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing deleted pipeline")
	return nil
}

func (c *PipelineJobController) patchPipelineJob(updated, original api.PipelineJob) (*api.PipelineJob, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineJobs(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
