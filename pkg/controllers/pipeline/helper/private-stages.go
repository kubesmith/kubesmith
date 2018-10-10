package helper

import (
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) getCurrentStageName() string {
	currentStage := p.pipeline.Status.StageIndex

	if currentStage == 0 {
		return ""
	} else if currentStage > len(p.pipeline.Spec.Stages) {
		return ""
	}

	return p.pipeline.Spec.Stages[currentStage-1]
}

func (p *PipelineHelper) getJobsForCurrentStage() []api.PipelineSpecJob {
	jobs := []api.PipelineSpecJob{}
	stageName := strings.ToLower(p.getCurrentStageName())

	if stageName == "" {
		return jobs
	}

	for _, job := range p.pipeline.Spec.Jobs {
		if strings.ToLower(job.Stage) == stageName {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func (p *PipelineHelper) advanceCurrentStageIndex() error {
	totalStages := len(p.pipeline.Spec.Stages)
	nextStage := p.pipeline.Status.StageIndex + 1

	if nextStage > totalStages {
		p.pipeline.Status.Phase = api.PipelinePhaseCompleted
	} else {
		p.pipeline.Status.StageIndex = nextStage
	}

	return p.patchPipeline()
}
