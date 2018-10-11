package helper

import (
	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) Execute() error {
	glog.V(1).Info("validating pipeline...")
	if err := p.Validate(); err != nil {
		glog.V(1).Info("pipeline was invalid; marking as failed")
		glog.Error(err)
		return p.SetPipelineStatus(api.PipelinePhaseFailed)
	}
	glog.V(1).Info("validated pipeline; processing current stage...")

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
	if err := p.validateEnvironmentVariables(p.pipeline.Spec.Environment); err != nil {
		return err
	}

	if err := p.validateTemplates(); err != nil {
		return err
	}

	if err := p.validateStages(); err != nil {
		return err
	}

	if err := p.validateJobs(); err != nil {
		return err
	}

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
