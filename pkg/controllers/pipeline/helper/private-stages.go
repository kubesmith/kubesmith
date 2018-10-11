package helper

import (
	"strings"

	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) processFirstStage() error {
	// mark the pipeline as "Running"
	glog.V(1).Info("setting pipeline to running...")
	if err := p.SetPipelineStatus(api.PipelinePhaseRunning); err != nil {
		glog.V(1).Info("could not set pipeline to running; carrying on")
		return err
	}

	// bootstrap minio server
	glog.V(1).Info("ensuring the minio server (and required resources) for this pipeline has been created...")
	if err := p.createMinioServer(); err != nil {
		glog.V(1).Info("could not ensure the minio server is running; setting the pipeline back to queued...")
		if err := p.SetPipelineStatus(api.PipelinePhaseQueued); err != nil {
			glog.V(1).Info("could not set the pipeline back to queued")
			return err
		}

		return err
	}

	// create the pipeline jobs in the system
	glog.V(1).Info("running pipeline jobs for current stage")
	if err := p.runPipelineJobsForCurrentStage(); err != nil {
		glog.V(1).Info("could not run pipeline jobs for current stage")
		return err
	}
	glog.V(1).Info("ran pipeline jobs for current stage!")

	return nil
}

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
