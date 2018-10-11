package helper

import (
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/pkg/errors"
)

func (p *PipelineHelper) runPipelineJobsForCurrentStage() error {
	jobsToRun := p.getJobsForCurrentStage()

	if len(jobsToRun) == 0 {
		return p.advanceCurrentStageIndex()
	}

	for _, job := range jobsToRun {
		job, err := p.expandJobTemplate(job)
		if err != nil {
			return errors.Wrap(err, "could not expand job template")
		}

		// todo: left off here
		_ = job
	}

	return nil
}

func (p *PipelineHelper) createPipelineJob() error {
	return nil
}

func (p *PipelineHelper) expandJobTemplate(job api.PipelineSpecJob) (*api.PipelineSpecJob, error) {
	newJob := job.DeepCopy()
	templates := p.pipeline.Spec.Templates
	envVars := []string{}
	artifacts := []api.PipelineSpecJobArtifact{}

	if len(newJob.Extends) == 0 {
		return newJob, nil
	}

	for _, templateName := range newJob.Extends {
		for _, template := range templates {
			if strings.ToLower(templateName) == strings.ToLower(template.Name) {
				newJob.Image = template.Image

				for _, env := range template.Environment {
					envVars = append(envVars, env)
				}

				newJob.AllowFailure = template.AllowFailure

				for _, artifact := range template.Artifacts {
					artifacts = append(artifacts, artifact)
				}

				for _, onlyOn := range template.OnlyOn {
					newJob.OnlyOn = append(newJob.OnlyOn, onlyOn)
				}
			}
		}
	}

	for _, env := range newJob.Environment {
		envVars = append(envVars, env)
	}

	for _, artifact := range newJob.Artifacts {
		artifacts = append(artifacts, artifact)
	}

	newJob.Environment = envVars
	newJob.Artifacts = artifacts

	return newJob, nil
}
