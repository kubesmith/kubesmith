package executor

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/validator"
)

func (p *PipelineExecutor) Validate() error {
	if err := validator.ValidateEnvironmentVariables(p._cachedPipeline.Spec.Environment); err != nil {
		return err
	}

	if err := validator.ValidateTemplates(p._cachedPipeline.Spec.Templates); err != nil {
		return err
	}

	if err := validator.ValidateStages(p._cachedPipeline.Spec.Stages); err != nil {
		return err
	}

	p.logger.Info("todo: expand jobs before validating them")
	return validator.ValidateJobs(
		p._cachedPipeline.Spec.Jobs,
		p._cachedPipeline.Spec.Stages,
		p._cachedPipeline.Spec.Templates,
	)
}

func (p *PipelineExecutor) Execute() error {
	switch p._cachedPipeline.Status.Phase {
	case api.PipelinePhaseEmpty, api.PipelinePhaseQueued:
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

func (p *PipelineExecutor) SetPipelineStatus(status api.PipelinePhase) error {
	switch status {
	case api.PipelinePhaseCompleted:
		p.Pipeline.Status.Phase = api.PipelinePhaseCompleted
		p.Pipeline.Status.StageIndex = len(p.Pipeline.Spec.Stages)
	case api.PipelinePhaseFailed:
		p.Pipeline.Status.Phase = api.PipelinePhaseFailed
		p.Pipeline.Status.StageIndex = len(p.Pipeline.Spec.Stages)
	case api.PipelinePhaseRunning:
		p.Pipeline.Status.Phase = api.PipelinePhaseRunning
		p.Pipeline.Status.StageIndex = 1
	case api.PipelinePhaseQueued:
		p.Pipeline.Status.Phase = api.PipelinePhaseQueued
		p.Pipeline.Status.StageIndex = 0
	}

	return p.patchPipeline()
}
