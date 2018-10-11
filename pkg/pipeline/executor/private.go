package executor

import (
	"encoding/json"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (p *PipelineExecutor) canRunAnotherPipeline() (bool, error) {
	pipelines, err := p.kubesmithClient.Pipelines(p.Pipeline.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines.Items {
		if pipeline.Status.Phase == api.PipelinePhaseRunning {
			currentlyRunning++
		}
	}

	if currentlyRunning < p.MaxRunningPipelines {
		return true, nil
	}

	return false, nil
}

// validate the pipeline before doing anything
// mark the pipeline as running
// bootstrap minio server
// 		if any issues, mark the pipeline as queued again
// wait for minio server to be "running"
// create the pipeline jobs in the system
func (p *PipelineExecutor) processQueuedPipeline() error {
	p.logger.Info("validating pipeline...")
	if err := p.Validate(); err != nil {
		p.logger.Info("could not validate pipeline")
		p.logger.Error(err)

		p.logger.Info("marking pipeline as failed")
		if err := p.SetPipelineStatus(api.PipelinePhaseFailed); err != nil {
			p.logger.Info("could not mark pipeline as failed")
			p.logger.Error(err)
		}

		return err
	}
	p.logger.Info("finished validating pipeline")

	return nil
}

func (p *PipelineExecutor) processRunningPipeline() error {
	p.logger.Info("todo: processing running pipeline")
	return nil
}

func (p *PipelineExecutor) processFinishedPipeline() error {
	p.logger.Info("todo: processing finished pipeline")
	return nil
}

func (p *PipelineExecutor) patchPipeline() error {
	p.Pipeline.Status.LastUpdated.Time = time.Now()

	origBytes, err := json.Marshal(p._cachedPipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling original pipeline")
	}

	updatedBytes, err := json.Marshal(p.Pipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling updated pipeline")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return errors.Wrap(err, "error creating json merge patch for pipeline")
	}

	res, err := p.kubesmithClient.Pipelines(p._cachedPipeline.Namespace).Patch(p._cachedPipeline.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return errors.Wrap(err, "error patching pipeline")
	}

	p._cachedPipeline = *res
	p.Pipeline = res.DeepCopy()
	return nil
}

func (p *PipelineExecutor) getCurrentStageName() string {
	currentStage := p._cachedPipeline.Status.StageIndex

	if currentStage == 0 {
		return ""
	} else if currentStage > len(p._cachedPipeline.Spec.Stages) {
		return ""
	}

	return p.Pipeline.Spec.Stages[currentStage-1]
}

func (p *PipelineExecutor) getJobsForCurrentStage() []api.PipelineSpecJob {
	jobs := []api.PipelineSpecJob{}
	stageName := strings.ToLower(p.getCurrentStageName())

	if stageName == "" {
		return jobs
	}

	for _, job := range p._cachedPipeline.Spec.Jobs {
		if strings.ToLower(job.Stage) == stageName {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func (p *PipelineExecutor) advanceCurrentStageIndex() error {
	totalStages := len(p.Pipeline.Spec.Stages)
	nextStage := p.Pipeline.Status.StageIndex + 1

	if nextStage > totalStages {
		p.Pipeline.Status.Phase = api.PipelinePhaseCompleted
	} else {
		p.Pipeline.Status.StageIndex = nextStage
	}

	return p.patchPipeline()
}

//
