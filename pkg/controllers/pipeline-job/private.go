package pipelinejob

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *PipelineJobController) processPipelineJob(action sync.SyncAction) error {
	cachedJob := action.GetObject().(api.PipelineJob)

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": cachedJob.GetName(),
		})

		if err := c.processDeletedPipelineJob(*cachedJob.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		job, err := c.pipelineJobLister.PipelineJobs(cachedJob.GetNamespace()).Get(cachedJob.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline job")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline job")
		}

		// create a new logger for this pipeline job's execution
		logger := c.logger.WithFields(logrus.Fields{
			"Name":  job.GetName(),
			"Phase": job.Status.Phase,
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
	job := *original.DeepCopy()

	logger.Info("checking if another pipeline job can be run")
	canRunAnotherPipelineJob, err := c.canRunAnotherPipelineJob(job)
	if err != nil {
		return errors.Wrap(err, "could not check if another pipeline job could be run")
	}

	if !canRunAnotherPipelineJob {
		logger.Info("cannot run another pipeline job")
		return nil
	}

	logger.Info("marking as running")
	job.SetPhaseToRunning()
	if _, err := c.patchPipelineJob(job, original); err != nil {
		return errors.Wrap(err, "could not mark as running")
	}

	logger.Info("marked as running")
	return nil
}

func (c *PipelineJobController) processRunningPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	err := c.ensureJobConfigMapIsScheduled(
		original.GetName(),
		original.GetNamespace(),
		c.getWrappedLabels(original),
		original.GetConfigMapData(),
		logger,
	)

	if err != nil {
		return errors.Wrap(err, "could not ensure job configmap is scheduled")
	}

	if err := c.ensureJobIsScheduled(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure job is scheduled")
	}

	return nil
}

func (c *PipelineJobController) processSuccessfulPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing successful pipeline job")
	return nil
}

func (c *PipelineJobController) processFailedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing failed pipeline job")
	return nil
}

func (c *PipelineJobController) processDeletedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("todo: processing deleted pipeline job")
	return nil
}

func (c *PipelineJobController) patchPipelineJob(updated, original api.PipelineJob) (*api.PipelineJob, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineJobs(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}

func (c *PipelineJobController) canRunAnotherPipelineJob(original api.PipelineJob) (bool, error) {
	jobs, err := c.pipelineJobLister.PipelineJobs(original.GetNamespace()).List(labels.Everything())
	if err != nil {
		return false, errors.Wrap(err, "could not list pipeline jobs")
	}

	currentlyRunning := 0
	for _, job := range jobs {
		if job.IsRunning() {
			currentlyRunning++
		}
	}

	if currentlyRunning < c.maxRunningPipelineJobs {
		return true, nil
	}

	return false, nil
}

func (c *PipelineJobController) ensureJobConfigMapIsScheduled(name, namespace string, labels, configMapData map[string]string, logger logrus.FieldLogger) error {
	logger.Info("ensuring job configmap is scheduled")
	if _, err := c.configMapLister.ConfigMaps(namespace).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("job configmap does not exist; scheduling")

			configMap := GetJobConfigMap(name, labels, configMapData)
			if _, err := c.kubeClient.CoreV1().ConfigMaps(namespace).Create(&configMap); err != nil {
				return errors.Wrap(err, "could not schedule job configmap")
			}

			logger.Info("job configmap was scheduled")
			return nil
		}

		return errors.Wrap(err, "could not get job configmap")
	}

	logger.Info("job configmap is scheduled")
	return nil
}

func (c *PipelineJobController) ensureJobIsScheduled(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("ensuring job is scheduled")

	// todo: left off here

	logger.Info("job is scheduled")
	return nil
}

func (c *PipelineJobController) getWrappedLabels(original api.PipelineJob) map[string]string {
	labels := original.GetLabels()
	labels["Controller"] = "PipelineJob"
	labels["PipelineJobName"] = original.GetName()
	labels["PipelineJobNamespace"] = original.GetNamespace()

	return labels
}
