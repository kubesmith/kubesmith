package pipelinejob

import (
	"fmt"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *PipelineJobController) processPipelineJob(action sync.SyncAction) error {
	pipelineJob := action.GetObject().(*batchv1.Job)
	if pipelineJob == nil {
		panic(errors.New("programmer error; pipeline job object was nil"))
	}

	tmpLogger := c.logger.WithFields(logrus.Fields{
		"JobName":      pipelineJob.GetName(),
		"JobNamespace": pipelineJob.GetNamespace(),
	})

	tmpLogger.Info("retrieving pipeline job...")
	pipelineJob, err := c.jobLister.Jobs(pipelineJob.GetNamespace()).Get(pipelineJob.GetName())
	if apierrors.IsNotFound(err) {
		tmpLogger.Error("pipeline job does not exist")
		return nil
	} else if err != nil {
		err = errors.Wrap(err, "error retrieving pipeline job")
		tmpLogger.Error(err)
		return err
	}
	tmpLogger.Info("retrieved pipeline job")

	// check to see if the correlated pipeline still exists
	tmpLogger.Info("retrieving pipeline from pipeline job...")
	pipeline, err := c.getPipelineFromJobLabels(pipelineJob)
	if err != nil {
		tmpLogger.Error(errors.Wrap(err, "could not retrieve pipeline from pipeline job"))
		return err
	}

	tmpLogger = tmpLogger.WithFields(logrus.Fields{
		"PipelineName":      pipeline.GetName(),
		"PipelineNamespace": pipeline.GetNamespace(),
	})
	tmpLogger.Info("retrieved pipeline from pipeline job")

	// check to see the outcome of the job
	if pipelineJob.Status.Failed == 1 {
		return c.processFailedPipelineJob(pipelineJob, *pipeline, tmpLogger.WithField("Status", "Failed"))
	} else if pipelineJob.Status.Succeeded == 1 {
		return c.processSucceededPipelineJob(pipelineJob, *pipeline, tmpLogger.WithField("Status", "Succeeded"))
	} else {
		tmpLogger.Info("skipping job since it hasn't been completed yet")
	}

	return nil
}

func (c *PipelineJobController) processFailedPipelineJob(job *batchv1.Job, pipeline api.Pipeline, logger logrus.FieldLogger) error {
	logger.Warn("todo: finish failed pipeline job")
	logger.Warn("todo: handle allowFailure for pipeline job")
	return nil
}

func (c *PipelineJobController) processSucceededPipelineJob(job *batchv1.Job, originalPipeline api.Pipeline, logger logrus.FieldLogger) error {
	jobsAreFinished, err := c.pipelineJobsAreFinishedForCurrentStage(originalPipeline, logger)
	if err != nil {
		logger.Error(err)
		return err
	}

	if !jobsAreFinished {
		logger.Info("skipping... job is not last in current stage")
		return nil
	}

	pipeline := *originalPipeline.DeepCopy()
	pipeline.AdvanceCurrentStage()
	if _, err := c.patchPipeline(pipeline, originalPipeline); err != nil {
		err = errors.Wrap(err, "could not update advanced pipeline")
		logger.Error(err)
		return err
	}

	return nil
}

func (c *PipelineJobController) pipelineJobsAreFinishedForCurrentStage(pipeline api.Pipeline, logger logrus.FieldLogger) (bool, error) {
	logger.Info("getting scheduled jobs...")
	scheduledJobs, err := c.getPipelineJobsForCurrentStage(pipeline)
	if err != nil {
		logger.Error(errors.Wrap(err, "could not get scheduled jobs for pipeline"))
		return false, err
	}
	logger.Info("got scheduled jobs for current pipeline stage...")

	totalScheduledJobs := len(scheduledJobs)
	totalPipelineJobs := len(pipeline.GetExpandedJobsForCurrentStage())

	if totalScheduledJobs == totalPipelineJobs {
		allJobsHaveFinished := true

		for _, job := range scheduledJobs {
			if job.Status.Failed == 0 && job.Status.Succeeded == 0 {
				allJobsHaveFinished = false
			}
		}

		return allJobsHaveFinished, nil
	}

	return false, nil
}

func (c *PipelineJobController) getPipelineJobsForCurrentStage(pipeline api.Pipeline) ([]*batchv1.Job, error) {
	allJobs, err := c.jobLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	expectedNamePrefix := fmt.Sprintf(
		"%s-stage-%d-job-",
		pipeline.GetResourcePrefix(),
		pipeline.GetStageIndex(),
	)

	jobs := []*batchv1.Job{}
	for _, job := range allJobs {
		if strings.HasPrefix(job.GetName(), expectedNamePrefix) == true {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (c *PipelineJobController) getPipelineFromJobLabels(job *batchv1.Job) (*api.Pipeline, error) {
	labels := job.GetLabels()
	pipeline, err := c.pipelineLister.Pipelines(labels["PipelineNamespace"]).Get(labels["PipelineName"])

	if err != nil {
		return nil, errors.Wrap(err, "could not fetch pipeline from job labels")
	}

	return pipeline, nil
}

func (c *PipelineJobController) jobIsPipelineJob(job *batchv1.Job) bool {
	// make sure the job has completed
	if job.Status.Failed == 0 && job.Status.Succeeded == 0 {
		return false
	}

	// now, check that the job has the labels we expect
	expectedLabels := []string{
		"PipelineName",
		"PipelineNamespace",
		"PipelineID",
	}
	actualLabels := job.GetLabels()

	for _, label := range expectedLabels {
		if _, exists := actualLabels[label]; !exists {
			return false
		}
	}

	// everything checks out, this seems like a valid pipeline job
	return true
}

func (c *PipelineJobController) patchPipeline(updated, original api.Pipeline) (*api.Pipeline, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.Pipelines(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
