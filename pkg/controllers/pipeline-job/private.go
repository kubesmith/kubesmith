package pipelinejob

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/utils"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

func (c *PipelineJobController) processPipelineJob(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	pipelineJob, err := c.jobLister.Jobs(ns).Get(name)
	if apierrors.IsNotFound(err) {
		glog.V(1).Info("unable to find pipeline job")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting pipeline job")
	}

	// check to see if the correlated pipeline still exists
	pipeline, err := c.getPipelineFromJobLabels(pipelineJob)
	if err != nil {
		return err
	}

	// check to see the outcome of the job
	if pipelineJob.Status.Failed == 1 {
		return c.processFailedPipelineJob(pipelineJob, *pipeline)
	} else if pipelineJob.Status.Succeeded == 1 {
		return c.processSucceededPipelineJob(pipelineJob, *pipeline)
	}

	return nil
}

func (c *PipelineJobController) processFailedPipelineJob(job *batchv1.Job, pipeline v1.Pipeline) error {
	c.logger.Warn("todo: handle allowFailure for pipeline job")

	updated := utils.SetPipelineToFailed(*pipeline.DeepCopy(), "job failed")

	if _, err := utils.PatchPipeline(pipeline, updated, c.kubesmithClient); err != nil {
		return errors.Wrap(err, "could not update pipeline phase to failed")
	}

	return nil
}

func (c *PipelineJobController) processSucceededPipelineJob(job *batchv1.Job, pipeline v1.Pipeline) error {
	c.logger.Warn("todo: handle succeeded pipeline job")
	return nil
}

func (c *PipelineJobController) getPipelineFromJobLabels(job *batchv1.Job) (*v1.Pipeline, error) {
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
		"PipelineMD5",
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
