package pipelinestage

import (
	"fmt"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *PipelineStageController) processPipelineStage(action sync.SyncAction) error {
	cachedStage := action.GetObject().(api.PipelineStage)

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger := c.logger.WithFields(logrus.Fields{
			"Name": cachedStage.GetName(),
		})

		if err := c.processDeletedPipelineStage(*cachedStage.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		stage, err := c.pipelineStageLister.PipelineStages(cachedStage.GetNamespace()).Get(cachedStage.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find pipeline stage")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline stage")
		}

		// create a new logger for this pipeline stage's execution
		logger := c.logger.WithFields(logrus.Fields{
			"Name":  stage.GetName(),
			"Phase": stage.Status.Phase,
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
	for index, job := range original.Spec.Jobs {
		name := fmt.Sprintf("%s-job-%d", original.GetName(), index+1)

		logger.Info("ensuring pipeline job is scheduled")
		if err := c.ensureJobIsScheduled(name, original, job, logger); err != nil {
			return errors.Wrap(err, "could not ensure pipeline job was scheduled")
		}

		logger.Info("pipeline job is scheduled")
	}

	logger.Info("all pipeline jobs are scheduled")
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
	// create a selector for listing resources associated to pipeline stage
	logger.Info("creating label selector for associated resources")
	labelSelector, err := c.getResourceLabelSelector(c.getWrappedLabels(original))
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
	logger.Info("retrieving pipeline jobs")
	jobs, err := c.pipelineJobLister.PipelineJobs(original.GetNamespace()).List(labelSelector)
	if err != nil {
		return errors.Wrap(err, "could not retrieve pipeline jobs")
	}
	logger.Info("retrieved pipeline jobs")

	// delete all of the pipeline jobs associated to the pipeline
	logger.Info("deleting pipeline jobs")
	for _, job := range jobs {
		if err := c.kubesmithClient.PipelineJobs(job.GetNamespace()).Delete(job.GetName(), &deleteOptions); err != nil {
			return errors.Wrapf(err, "could not delete pipeline job: %s/%s", job.GetName(), job.GetNamespace())
		}
	}

	logger.Info("deleted pipeline jobs")
	return nil
}

func (c *PipelineStageController) patchPipelineStage(updated, original api.PipelineStage) (*api.PipelineStage, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineStages(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}

func (c *PipelineStageController) getWrappedLabels(original api.PipelineStage) map[string]string {
	labels := original.GetLabels()
	labels[api.GetLabelKey("Controller")] = "PipelineStage"
	labels[api.GetLabelKey("PipelineStageName")] = original.GetName()
	labels[api.GetLabelKey("PipelineStageNamespace")] = original.GetNamespace()

	return labels
}

func (c *PipelineStageController) ensureJobIsScheduled(
	name string,
	original api.PipelineStage,
	job api.PipelineJobSpecJob,
	logger logrus.FieldLogger,
) error {
	if _, err := c.pipelineJobLister.PipelineJobs(original.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("pipeline job does not exist; scheduling")

			job := GetPipelineJob(
				name,
				c.getWrappedLabels(original),
				original.Spec.Workspace.Repo,
				original.Spec.Workspace.Storage,
				job,
			)
			if _, err := c.kubesmithClient.PipelineJobs(original.GetNamespace()).Create(&job); err != nil {
				return errors.Wrap(err, "could not schedule pipeline job")
			}

			logger.Info("pipeline job was scheduled")
			return nil
		}

		return errors.Wrap(err, "could not get pipeline stage")
	}

	logger.Info("pipeline job is scheduled")
	return nil
}

func (c *PipelineStageController) cleanupJob(name string) error {
	return nil
}

func (c *PipelineStageController) getResourceLabelSelector(resourceLabels map[string]string) (labels.Selector, error) {
	selector := labels.NewSelector()

	for key, value := range resourceLabels {
		req, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return nil, errors.Wrap(err, "could not create label requirement")
		}

		selector.Add(*req)
	}

	return selector, nil
}
