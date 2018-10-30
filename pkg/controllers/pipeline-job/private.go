package pipelinejob

import (
	"fmt"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *PipelineJobController) processPipelineJob(action sync.SyncAction) error {
	cachedJob := action.GetObject().(api.PipelineJob)

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": cachedJob.GetName(),
		})

		return c.processDeletedPipelineJob(*cachedJob.DeepCopy(), logger)
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
			return c.processEmptyPhasePipelineJob(*job.DeepCopy(), logger)
		} else if job.IsQueued() {
			return c.processQueuedPipelineJob(*job.DeepCopy(), logger)
		} else if job.IsRunning() {
			return c.processRunningPipelineJob(*job.DeepCopy(), logger)
		} else if job.HasSucceeded() {
			return c.processSuccessfulPipelineJob(*job.DeepCopy(), logger)
		} else if job.HasFailed() {
			return c.processFailedPipelineJob(*job.DeepCopy(), logger)
		}
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
	if err := c.ensureJobConfigMapIsScheduled(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure job configmap is scheduled")
	}

	if err := c.ensureJobIsScheduled(original, logger); err != nil {
		return errors.Wrap(err, "could not ensure job is scheduled")
	}

	return nil
}

func (c *PipelineJobController) processSuccessfulPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("fetching associated pipeline stage")
	pipelineStage, err := c.getAssociatedPipelineStage(original)
	if err != nil {
		return err
	}
	logger.Info("fetched associated pipeline stage")

	if pipelineStage.HasSucceeded() || pipelineStage.HasFailed() {
		logger.Info("pipeline stage has already completed; skipping")
		return nil
	}

	// stuff an entry into the pipeline stage for this job
	logger.Info("adding pipeline job completion to pipeline stage")
	updatedPipelineStage := c.markPipelineJobAsCompleted(original, *pipelineStage.DeepCopy(), api.PhaseSucceeded)

	if _, err := c.patchPipelineStage(updatedPipelineStage, *pipelineStage); err != nil {
		return errors.Wrap(err, "could not add pipeline job completion to pipeline stage")
	}

	logger.Info("added pipeline job completion to pipeline stage")
	return nil
}

func (c *PipelineJobController) markPipelineJobAsCompleted(pipelineJob api.PipelineJob, pipelineStage api.PipelineStage, phase api.Phase) api.PipelineStage {
	key := fmt.Sprintf("%s/%s", pipelineJob.GetNamespace(), pipelineJob.GetName())

	if pipelineStage.Status.CompletedPipelineJobs == nil {
		pipelineStage.Status.CompletedPipelineJobs = map[string]string{}
	}

	pipelineStage.Status.CompletedPipelineJobs[key] = string(phase)
	return pipelineStage
}

func (c *PipelineJobController) processFailedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	if original.Spec.Job.AllowFailure == true {
		return c.processSuccessfulPipelineJob(original, logger)
	}

	logger.Info("fetching associated pipeline stage")
	pipelineStage, err := c.getAssociatedPipelineStage(original)
	if err != nil {
		return err
	}
	logger.Info("fetched associated pipeline stage")

	if pipelineStage.HasSucceeded() || pipelineStage.HasFailed() {
		logger.Info("pipeline stage has already completed; skipping")
		return nil
	}

	logger.Info("marking pipeline stage as failed")
	updatedPipelineStage := c.markPipelineJobAsCompleted(original, *pipelineStage.DeepCopy(), api.PhaseFailed)

	// todo: improve this failure reason
	updatedPipelineStage.SetPhaseToFailed("pipeline job failed")

	if _, err := c.patchPipelineStage(updatedPipelineStage, *pipelineStage); err != nil {
		return errors.Wrap(err, "could not mark pipeline stage as failed")
	}

	logger.Info("marked pipeline stage as failed")
	return nil
}

func (c *PipelineJobController) getAssociatedPipelineStage(original api.PipelineJob) (*api.PipelineStage, error) {
	name, err := c.getLabelByKey(original, "PipelineStageName")
	if err != nil {
		return nil, err
	}

	namespace, err := c.getLabelByKey(original, "PipelineStageNamespace")
	if err != nil {
		return nil, err
	}

	return c.pipelineStageLister.PipelineStages(namespace).Get(name)
}

func (c *PipelineJobController) getLabelByKey(original api.PipelineJob, key string) (string, error) {
	labels := original.GetLabels()

	if value, ok := labels[api.GetLabelKey(key)]; ok {
		return value, nil
	}

	return "", errors.New("could not find pipeline job label")
}

func (c *PipelineJobController) processDeletedPipelineJob(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("cleaning up pipeline job")

	// create a selector for listing resources associated to pipelines
	labelSelector := c.getResourceLabelSelector(c.getWrappedLabels(original))

	// create the delete options that can help clean everything up
	propagationPolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	if err := c.deleteAssociatedConfigMaps(original, labelSelector, deleteOptions, logger); err != nil {
		return err
	}

	if err := c.deleteAssociatedJobs(original, labelSelector, deleteOptions, logger); err != nil {
		return err
	}

	logger.Info("pipeline cleaned up")
	return nil
}

func (c *PipelineJobController) deleteAssociatedConfigMaps(
	original api.PipelineJob,
	labelSelector labels.Selector,
	deleteOptions metav1.DeleteOptions,
	logger logrus.FieldLogger,
) error {
	logger.Info("retrieving configmaps")
	configMaps, err := c.configMapLister.ConfigMaps(original.GetNamespace()).List(labelSelector)
	if err != nil {
		return errors.Wrap(err, "could not retrieve configmaps")
	}

	logger.Info("retrieved configmaps")

	// delete all of the configmaps associated to the pipeline job
	logger.Info("deleting configmaps")
	for _, configMap := range configMaps {
		logger.WithFields(logrus.Fields{
			"ConfigMapName":      configMap.GetName(),
			"ConfigMapNamespace": configMap.GetNamespace(),
		}).Info("deleting configmap")

		if err := c.kubeClient.CoreV1().ConfigMaps(configMap.GetNamespace()).Delete(configMap.GetName(), &deleteOptions); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}

			return errors.Wrapf(err, "could not delete configmap: %s/%s", configMap.GetNamespace(), configMap.GetName())
		}
	}

	logger.Info("deleted configmaps")
	return nil
}

func (c *PipelineJobController) deleteAssociatedJobs(
	original api.PipelineJob,
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

	// delete all of the configmaps associated to the pipeline job
	logger.Info("deleting jobs")
	for _, job := range jobs {
		logger.WithFields(logrus.Fields{
			"JobName":      job.GetName(),
			"JobNamespace": job.GetNamespace(),
		}).Info("deleting job")

		if err := c.kubeClient.BatchV1().Jobs(original.GetNamespace()).Delete(job.GetName(), &deleteOptions); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}

			return errors.Wrapf(err, "could not delete job: %s/%s", job.GetNamespace(), job.GetName())
		}
	}

	logger.Info("deleted jobs")
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

func (c *PipelineJobController) ensureJobConfigMapIsScheduled(original api.PipelineJob, logger logrus.FieldLogger) error {
	logger.Info("ensuring job configmap is scheduled")
	if _, err := c.configMapLister.ConfigMaps(original.GetNamespace()).Get(original.GetName()); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("job configmap does not exist; scheduling")

			configMap := GetJobConfigMap(
				original.GetName(),
				c.getWrappedLabels(original),
				original.GetConfigMapData(),
			)

			if _, err := c.kubeClient.CoreV1().ConfigMaps(original.GetNamespace()).Create(&configMap); err != nil {
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
	if _, err := c.jobLister.Jobs(original.GetNamespace()).Get(original.GetName()); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("job does not exist; scheduling")

			command := original.Spec.Job.Command
			args := original.Spec.Job.Args

			if len(original.Spec.Job.Runner) > 0 {
				// runner trumps all
				command = []string{"/bin/sh", "-x", "/kubesmith/scripts/pipeline-script.sh"}
				args = []string{}
			}

			job := GetJob(
				original.GetName(),
				original.Spec.Job.Image,
				original.Spec.Workspace.Path,
				command,
				args,
				original.Spec.Workspace.Storage.S3,
				original.Spec.Job.Environment,
				c.getWrappedLabels(original),
			)

			if _, err := c.kubeClient.BatchV1().Jobs(original.GetNamespace()).Create(&job); err != nil {
				return errors.Wrap(err, "could not schedule job")
			}

			logger.Info("job is scheduled")
			return nil
		}

		return errors.Wrap(err, "could not get job")
	}

	logger.Info("job is scheduled")
	return nil
}

func (c *PipelineJobController) getWrappedLabels(original api.PipelineJob) map[string]string {
	labels := original.GetLabels()
	labels[api.GetLabelKey("Controller")] = "PipelineJob"
	labels[api.GetLabelKey("PipelineJobName")] = original.GetName()
	labels[api.GetLabelKey("PipelineJobNamespace")] = original.GetNamespace()

	return labels
}

func (c *PipelineJobController) getResourceLabelSelector(resourceLabels map[string]string) labels.Selector {
	set := labels.Set{}

	for key, value := range resourceLabels {
		set[key] = value
	}

	return labels.SelectorFromSet(set)
}

func (c *PipelineJobController) patchPipelineStage(updated, original api.PipelineStage) (*api.PipelineStage, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineStages(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
