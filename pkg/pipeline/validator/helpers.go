package validator

import (
	"errors"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func ValidateEnvironmentVariables(vars []string) error {
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

func ValidateArtifacts(artifacts []api.PipelineSpecJobArtifact) error {
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

func ValidateTemplates(templates []api.PipelineSpecJobTemplate) error {
	for _, template := range templates {
		if err := ValidateEnvironmentVariables(template.Environment); err != nil {
			return err
		}

		if err := ValidateArtifacts(template.Artifacts); err != nil {
			return err
		}
	}

	return nil
}

func ValidateStages(stages []string, jobs []api.PipelineSpecJob) error {
	if len(stages) == 0 {
		return errors.New("stages must include at least one value")
	}

	for _, stage := range stages {
		stage = strings.ToLower(stage)
		hasJobs := false

		for _, job := range jobs {
			if stage == strings.ToLower(job.Stage) {
				hasJobs = true
			}
		}

		if !hasJobs {
			return errors.New("each stage must have jobs associated to it")
		}
	}

	return nil
}

func ValidateJobs(jobs []api.PipelineSpecJob, stages []string, templates []api.PipelineSpecJobTemplate) error {
	for _, job := range jobs {
		// check the job has a valid name
		if job.Name == "" {
			return errors.New("job name must not be empty")
		}

		// check the job has an image specified
		if job.Image == "" {
			return errors.New("job image must not be empty")
		}

		// check the job has a stage specified
		if job.Stage == "" {
			return errors.New("job stage must be specified")
		}

		// check that the stage for this job is specified
		foundStage := false
		for _, stage := range stages {
			if strings.ToLower(stage) == strings.ToLower(job.Stage) {
				foundStage = true
			}
		}

		if !foundStage {
			return errors.New("job stage must be specified as a valid stage")
		}

		// check that all of the declared extends exist
		if len(job.Extends) > 0 {
			for _, extend := range job.Extends {
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

		// check the job to make sure it has command/args OR a runner
		hasCommands := len(job.Command) > 0 || len(job.Args) > 0
		hasRunner := len(job.Runner) > 0

		if hasCommands && hasRunner {
			return errors.New("job must have either command/args or runner specified; not both")
		}

		// check the job has valid environment variables
		if err := ValidateEnvironmentVariables(job.Environment); err != nil {
			return err
		}

		// check that the artifacts are valid
		if err := ValidateArtifacts(job.Artifacts); err != nil {
			return err
		}
	}

	return nil
}
