package helper

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) Execute() error {
	if err := p.Validate(); err != nil {
		return err
	}

	switch p.pipeline.Status.Phase {
	case api.PipelinePhaseEmpty, api.PipelinePhaseQueued:
		if err := p.processFirstStage(); err != nil {
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

func (p *PipelineHelper) Validate() error {
	return nil
}

func (p *PipelineHelper) SetPipelineStatus(status api.PipelinePhase) error {
	switch status {
	case api.PipelinePhaseCompleted:
		p.pipeline.Status.Phase = api.PipelinePhaseCompleted
		p.pipeline.Status.StageIndex = len(p.pipeline.Spec.Stages)
	case api.PipelinePhaseFailed:
		p.pipeline.Status.Phase = api.PipelinePhaseFailed
		p.pipeline.Status.StageIndex = len(p.pipeline.Spec.Stages)
	case api.PipelinePhaseRunning:
		p.pipeline.Status.Phase = api.PipelinePhaseRunning
		p.pipeline.Status.StageIndex = 1
	case api.PipelinePhaseQueued:
		p.pipeline.Status.Phase = api.PipelinePhaseQueued
		p.pipeline.Status.StageIndex = 0
	}

	return p.patchPipeline()
}
