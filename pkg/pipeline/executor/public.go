package executor

import (
	"fmt"
	"time"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/utils"
	"github.com/kubesmith/kubesmith/pkg/pipeline/validator"
)

func (p *PipelineExecutor) Validate() error {
	stages := p._cachedPipeline.Spec.Stages
	templates := p._cachedPipeline.Spec.Templates
	jobs := p._cachedPipeline.Spec.Jobs

	if err := validator.ValidateEnvironmentVariables(p._cachedPipeline.Spec.Environment); err != nil {
		return err
	}

	if err := validator.ValidateTemplates(templates); err != nil {
		return err
	}

	if err := validator.ValidateStages(stages, jobs); err != nil {
		return err
	}

	// Why do we expand all of the jobs before validating them?
	// Because it's possible that a pipeline job doesn't have a resource specified
	// but maybe the same pipeline job specifies an extension which has that
	// resource declared.
	expandedJobs := []api.PipelineSpecJob{}
	for _, job := range expandedJobs {
		expandedJobs = append(expandedJobs, *p.expandPipelineJob(job))
	}

	// lastly, validate the jobs
	return validator.ValidateJobs(expandedJobs, stages, templates)
}

func (p *PipelineExecutor) Execute() error {
	switch p._cachedPipeline.Status.Phase {
	case api.PipelinePhaseEmpty:
		if err := p.processEmptyPhasePipeline(); err != nil {
			return err
		}
	case api.PipelinePhaseQueued:
		if err := p.processQueuedPipeline(); err != nil {
			return err
		}
	case api.PipelinePhaseRunning:
		if err := p.processRunningPipeline(); err != nil {
			return err
		}
	case api.PipelinePhaseCompleted, api.PipelinePhaseFailed:
		if err := p.processFinishedPipeline(); err != nil {
			return err
		}
	}

	return nil
}

func (p *PipelineExecutor) GetNamespace() string {
	return p._cachedPipeline.Namespace
}

func (p *PipelineExecutor) SetPipelineToQueued() error {
	p.Pipeline.Status.StageIndex = 0
	p.Pipeline.Status.Phase = api.PipelinePhaseQueued

	return p.patchPipeline()
}

func (p *PipelineExecutor) SetPipelineToCompleted() error {
	p.Pipeline.Status.StageIndex = len(p.Pipeline.Spec.Stages)
	p.Pipeline.Status.Phase = api.PipelinePhaseCompleted
	p.Pipeline.Status.EndTime.Time = time.Now()

	return p.patchPipeline()
}

func (p *PipelineExecutor) SetPipelineToRunning() error {
	p.Pipeline.Status.StageIndex = 1
	p.Pipeline.Status.Phase = api.PipelinePhaseRunning
	p.Pipeline.Status.StartTime.Time = time.Now()

	return p.patchPipeline()
}

func (p *PipelineExecutor) SetPipelineToFailed(reason string) error {
	p.Pipeline.Status.StageIndex = len(p.Pipeline.Spec.Stages)
	p.Pipeline.Status.Phase = api.PipelinePhaseFailed
	p.Pipeline.Status.EndTime.Time = time.Now()
	p.Pipeline.Status.FailureReason = reason

	return p.patchPipeline()
}

func (p *PipelineExecutor) GetResourcePrefix() string {
	md5 := utils.GetPipelineMD5(p._cachedPipeline.Name, p._cachedPipeline.Namespace)

	return fmt.Sprintf("pipeline-%s", md5)
}

func (p *PipelineExecutor) GetResourceLabels() map[string]string {
	labels := map[string]string{
		"PipelineName":      p._cachedPipeline.Name,
		"PipelineNamespace": p._cachedPipeline.Namespace,
		"PipelineMD5":       utils.GetPipelineMD5(p._cachedPipeline.Name, p._cachedPipeline.Namespace),
	}

	return labels
}
