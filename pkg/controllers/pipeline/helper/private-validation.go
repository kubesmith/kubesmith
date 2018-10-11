package helper

import (
	"errors"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func (p *PipelineHelper) validateEnvironmentVariables(vars []string) error {
	for _, env := range vars {
		parts := strings.Split(env, "=")
		total := len(parts)

		if total == 1 {
			return errors.New("environment variables must contain a value")
		} else if total > 2 {
			return errors.New("environment variables must only contain a single key/value pair")
		}
	}

	return nil
}

func (p *PipelineHelper) validateArtifacts(artifacts []api.PipelineSpecJobArtifact) error {
	for _, artifact := range artifacts {
		if artifact.Name == "" {
			return errors.New("artifact name must not be empty")
		}

		switch artifact.When {
		case api.ArtifactEventOnSuccess, api.ArtifactEventOnFailure:
			break
		default:
			return errors.New("artifact when must be a valid event")
		}

		if len(artifact.Paths) == 0 {
			return errors.New("artifact paths must contain at least one path")
		}
	}

	return nil
}

func (p *PipelineHelper) validateTemplates() error {
	for _, template := range p.pipeline.Spec.Templates {
		if err := p.validateEnvironmentVariables(template.Environment); err != nil {
			return err
		}

		if err := p.validateArtifacts(template.Artifacts); err != nil {
			return err
		}
	}

	return nil
}

func (p *PipelineHelper) validateStages() error {
	if len(p.pipeline.Spec.Stages) == 0 {
		return errors.New("stages must include at least one value")
	}

	return nil
}

func (p *PipelineHelper) validateJobs() error {
	stages := p.pipeline.Spec.Stages
	templates := p.pipeline.Spec.Templates

	for _, job := range p.pipeline.Spec.Jobs {
		newJob := p.expandJobTemplate(job)

		// check the job has a valid name
		if newJob.Name == "" {
			return errors.New("job name must not be empty")
		}

		// check the job has an image specified
		if newJob.Image == "" {
			return errors.New("job image must not be empty")
		}

		// todo: validate the image pull secret (check that it exists)

		// check the job has a stage specified
		if newJob.Stage == "" {
			return errors.New("job stage must be specified")
		}

		// check that the stage for this job is specified
		foundStage := false
		for _, stage := range stages {
			if strings.ToLower(stage) == strings.ToLower(newJob.Stage) {
				foundStage = true
			}
		}

		if !foundStage {
			return errors.New("job stage must be specified as a valid stage")
		}

		// check that all of the declared extends exist
		if len(newJob.Extends) > 0 {
			for _, extend := range newJob.Extends {
				foundExtend := false

				for _, extension := range templates {
					if strings.ToLower(extend) == strings.ToLower(extension.Name) {
						foundExtend = true
					}
				}

				if !foundExtend {
					return errors.New("invalid job extension specified")
				}
			}
		}

		// check the job has valid environment variables
		if err := p.validateEnvironmentVariables(newJob.Environment); err != nil {
			return err
		}

		// check that the commands are valid
		if len(newJob.Commands) > 0 {
			for _, command := range newJob.Commands {
				if command == "" {
					return errors.New("job command entry must not be empty")
				}
			}
		}

		// check that the artifacts are valid
		if err := p.validateArtifacts(newJob.Artifacts); err != nil {
			return err
		}
	}

	return nil
}
