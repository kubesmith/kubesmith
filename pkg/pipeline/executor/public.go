package executor

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/validator"
)

func (p *PipelineExecutor) GetCachedPipeline() api.Pipeline {
	return *p._cachedPipeline.DeepCopy()
}

func (p *PipelineExecutor) Validate() error {
	pl := p.GetCachedPipeline()

	if err := validator.ValidateEnvironmentVariables(pl.GetEnvironment()); err != nil {
		return err
	}

	if err := validator.ValidateTemplates(pl.GetTemplates()); err != nil {
		return err
	}

	if err := validator.ValidateStages(pl.GetStages(), pl.GetJobs()); err != nil {
		return err
	}

	// lastly, validate the jobs
	return validator.ValidateJobs(pl.GetExpandedJobs(), pl.GetStages(), pl.GetTemplates())
}

func (p *PipelineExecutor) Execute() error {
	pl := p.GetCachedPipeline()

	switch pl.GetCurrentPhase() {
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
