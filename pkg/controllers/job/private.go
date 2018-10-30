package job

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
)

func (c *JobController) processJob(action sync.SyncAction) error {
	cachedJob := action.GetObject().(batchv1.Job)
	job, err := c.jobLister.Jobs(cachedJob.GetNamespace()).Get(cachedJob.GetName())
	if err != nil {
		return errors.Wrap(err, "error getting job")
	}

	// create a new logger for this pipeline's execution
	logger := c.logger.WithFields(logrus.Fields{
		"Name": job.GetName(),
	})

	if job.Status.Succeeded == 1 {
		return c.processSuccessfulJob(*job.DeepCopy(), logger)
	} else if job.Status.Failed == 1 {
		return c.processFailedJob(*job.DeepCopy(), logger)
	}

	return nil
}

func (c *JobController) processSuccessfulJob(original batchv1.Job, logger logrus.FieldLogger) error {
	logger.Info("fetching associated pipeline job")
	pipelineJob, err := c.getAssociatedPipelineJob(original)
	if err != nil {
		return err
	}
	logger.Info("fetched associated pipeline job")

	if pipelineJob.HasSucceeded() || pipelineJob.HasFailed() {
		logger.Info("pipeline job has already completed; skipping")
		return nil
	}

	logger.Info("updating associated pipeline job")
	updatedPipelineJob := *pipelineJob.DeepCopy()
	updatedPipelineJob.SetPhaseToSucceeded()

	if _, err := c.patchPipelineJob(updatedPipelineJob, *pipelineJob); err != nil {
		return errors.Wrap(err, "could not mark pipeline job as successful")
	}

	logger.Info("updated associated pipeline job")
	return nil
}

func (c *JobController) processFailedJob(original batchv1.Job, logger logrus.FieldLogger) error {
	logger.Info("fetching associated pipeline job")
	pipelineJob, err := c.getAssociatedPipelineJob(original)
	if err != nil {
		return err
	}
	logger.Info("fetched associated pipeline job")

	if pipelineJob.HasSucceeded() || pipelineJob.HasFailed() {
		logger.Info("pipeline job has already completed; skipping")
		return nil
	}

	logger.Info("updating associated pipeline job")
	updatedPipelineJob := *pipelineJob.DeepCopy()

	// todo: improve this failure reason
	updatedPipelineJob.SetPhaseToFailed("job failed")

	if _, err := c.patchPipelineJob(updatedPipelineJob, *pipelineJob); err != nil {
		return errors.Wrap(err, "could not mark pipeline job as a failure")
	}

	logger.Info("updated associated pipeline job")
	return nil
}

func (c *JobController) getAssociatedPipelineJob(original batchv1.Job) (*api.PipelineJob, error) {
	name, err := c.getPipelineJobName(original)
	if err != nil {
		return nil, err
	}

	namespace, err := c.getPipelineJobNamespace(original)
	if err != nil {
		return nil, err
	}

	return c.pipelineJobLister.PipelineJobs(namespace).Get(name)
}

func (c *JobController) getPipelineJobName(original batchv1.Job) (string, error) {
	labels := original.GetLabels()

	if value, ok := labels[api.GetLabelKey("PipelineJobName")]; ok {
		return value, nil
	}

	return "", errors.New("could not find pipeline job name label")
}

func (c *JobController) getPipelineJobNamespace(original batchv1.Job) (string, error) {
	labels := original.GetLabels()

	if value, ok := labels[api.GetLabelKey("PipelineJobNamespace")]; ok {
		return value, nil
	}

	return "", errors.New("could not find pipeline job namespace label")
}

func (c *JobController) jobHasWork(job *batchv1.Job) bool {
	isActive := job.Status.Active == 1
	hasSucceeded := job.Status.Succeeded == 1
	hasFailed := job.Status.Failed == 1
	labels := job.GetLabels()
	_, hasLabel := labels[api.GetLabelKey("PipelineJobName")]

	return !isActive && (hasSucceeded || hasFailed) && hasLabel
}

func (c *JobController) patchPipelineJob(updated, original api.PipelineJob) (*api.PipelineJob, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.PipelineJobs(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
