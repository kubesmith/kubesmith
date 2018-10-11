package helper

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) runPipelineJobsForCurrentStage() error {
	jobsToRun := p.getJobsForCurrentStage()

	if len(jobsToRun) == 0 {
		return p.advanceCurrentStageIndex()
	}

	for _, job := range jobsToRun {
		newJob := p.expandJobTemplate(job)

		// todo: left off here
		spew.Dump(newJob)
	}

	return nil
}

func (p *PipelineHelper) createPipelineJob() error {
	return nil
}

func (p *PipelineHelper) expandJobTemplate(job api.PipelineSpecJob) *api.PipelineSpecJob {
	newJob := job.DeepCopy()
	templates := p.pipeline.Spec.Templates
	envVars := []string{}
	artifacts := []api.PipelineSpecJobArtifact{}

	if len(newJob.Extends) == 0 {
		return newJob
	}

	for _, env := range p.pipeline.Spec.Environment {
		envVars = append(envVars, env)
	}

	for _, templateName := range newJob.Extends {
		for _, template := range templates {
			if strings.ToLower(templateName) == strings.ToLower(template.Name) {
				if template.Image != "" {
					newJob.Image = template.Image
				}

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

	return newJob
}
