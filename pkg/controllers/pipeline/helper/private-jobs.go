package helper

func (p *PipelineHelper) runPipelineJobsForCurrentStage() error {
	jobsToRun := p.getJobsForCurrentStage()

	if len(jobsToRun) == 0 {
		return p.advanceCurrentStageIndex()
	}

	//

	return nil
}

func (p *PipelineHelper) createPipelineJob() error {
	return nil
}
