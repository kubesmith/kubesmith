package executor

import (
	"encoding/json"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (p *PipelineExecutor) canRunAnotherPipeline() (bool, error) {
	pipelines, err := p.kubesmithClient.Pipelines(p.Pipeline.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines.Items {
		if pipeline.Status.Phase == api.PipelinePhaseRunning {
			currentlyRunning++
		}
	}

	if currentlyRunning < p.MaxRunningPipelines {
		return true, nil
	}

	return false, nil
}

func (p *PipelineExecutor) processEmptyPhasePipeline() error {
	// first, validate the pipeline before doing anything
	p.logger.Info("validating pipeline...")
	if err := p.Validate(); err != nil {
		p.logger.Error("could not validate pipeline")
		p.logger.Error(err)

		p.logger.Info("marking pipeline as failed")
		if err := p.SetPipelineToFailed(err.Error()); err != nil {
			p.logger.Error("could not mark pipeline as failed")
			p.logger.Error(err)
		}

		return err
	}
	p.logger.Info("finished validating pipeline")

	// lastly, set the pipeline status to queued
	p.logger.Info("marking pipeline as queued...")
	if err := p.SetPipelineToQueued(); err != nil {
		p.logger.Error("could not set pipeline to running")
		p.logger.Error(err)

		return err
	}
	p.logger.Info("marked pipeline as queued")

	return nil
}

func (p *PipelineExecutor) processQueuedPipeline() error {
	// first, validate the pipeline before doing anything
	p.logger.Info("validating pipeline...")
	if err := p.Validate(); err != nil {
		p.logger.Error("could not validate pipeline")
		p.logger.Error(err)

		p.logger.Info("marking pipeline as failed")
		if err := p.SetPipelineToFailed(err.Error()); err != nil {
			p.logger.Error("could not mark pipeline as failed")
			p.logger.Error(err)
		}

		return err
	}
	p.logger.Info("finished validating pipeline")

	// check to see if we can run another pipeline in this namespace
	p.logger.Info("checking to see if we can run another pipeline in this namespace")
	canRunAnotherPipeline, err := p.canRunAnotherPipeline()
	if err != nil {
		p.logger.Error("could not check to see if we could run another pipeline")
		p.logger.Error(err)
		return err
	}

	if !canRunAnotherPipeline {
		// don't do anything, disregard this queue update
		p.logger.Warn("cannot run another pipeline in this namespace")
		return nil
	}

	// next, mark the pipeline as running
	p.logger.Info("marking pipeline as running...")
	if err := p.SetPipelineToRunning(); err != nil {
		p.logger.Error("could not set pipeline to running")
		p.logger.Error(err)

		return err
	}
	p.logger.Info("marked pipeline as running")

	return nil
}

// make sure minio server is bootstrapped
// 		if any issues, mark pipeline as failed
//		if minio server did not exist, create it
// wait for minio server to be "ready"
// create the pipeline jobs for current stage in the system
func (p *PipelineExecutor) processRunningPipeline() error {
	p.logger.Info("todo: processing running pipeline")
	return nil
}

// cleanup all resources associated with the pipeline
func (p *PipelineExecutor) processFinishedPipeline() error {
	p.logger.Info("todo: processing finished pipeline")
	return nil
}

func (p *PipelineExecutor) patchPipeline() error {
	p.Pipeline.Status.LastUpdated.Time = time.Now()

	origBytes, err := json.Marshal(p._cachedPipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling original pipeline")
	}

	updatedBytes, err := json.Marshal(p.Pipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling updated pipeline")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return errors.Wrap(err, "error creating json merge patch for pipeline")
	}

	res, err := p.kubesmithClient.Pipelines(p._cachedPipeline.Namespace).Patch(p._cachedPipeline.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return errors.Wrap(err, "error patching pipeline")
	}

	p._cachedPipeline = *res
	p.Pipeline = res.DeepCopy()
	return nil
}

func (p *PipelineExecutor) getCurrentStageName() string {
	currentStage := p._cachedPipeline.Status.StageIndex

	if currentStage == 0 {
		return ""
	} else if currentStage > len(p._cachedPipeline.Spec.Stages) {
		return ""
	}

	return p.Pipeline.Spec.Stages[currentStage-1]
}

func (p *PipelineExecutor) getJobsForCurrentStage() []api.PipelineSpecJob {
	jobs := []api.PipelineSpecJob{}
	stageName := strings.ToLower(p.getCurrentStageName())

	if stageName == "" {
		return jobs
	}

	for _, job := range p._cachedPipeline.Spec.Jobs {
		if strings.ToLower(job.Stage) == stageName {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func (p *PipelineExecutor) advanceCurrentStageIndex() error {
	totalStages := len(p.Pipeline.Spec.Stages)
	nextStage := p.Pipeline.Status.StageIndex + 1

	if nextStage > totalStages {
		p.Pipeline.Status.Phase = api.PipelinePhaseCompleted
	} else {
		p.Pipeline.Status.StageIndex = nextStage
	}

	return p.patchPipeline()
}

func (p *PipelineExecutor) expandJobPipeline(oldJob api.PipelineSpecJob) *api.PipelineSpecJob {
	job := oldJob.DeepCopy()
	envVars := []string{}
	artifacts := []api.PipelineSpecJobArtifact{}

	// if this job doesn't extend anything, we're done
	if len(job.Extends) == 0 {
		return job
	}

	// loop through the pipeline's global environment variables and add them first
	for _, env := range p._cachedPipeline.Spec.Environment {
		envVars = append(envVars, env)
	}

	// loop through the job's specified extensions and use each extension to mutate
	// the job in the order they were specified
	for _, templateName := range job.Extends {
		template, _ := p.getTemplateByName(templateName)

		// if there was no template by that name, keep moving on
		if template == nil {
			continue
		}

		// if the template has an image specified, overwrite the job's image
		if template.Image != "" {
			job.Image = template.Image
		}

		// add the environment variables from this template
		for _, env := range template.Environment {
			envVars = append(envVars, env)
		}

		// add the artifacts from this template (if any were specified)
		for _, artifact := range template.Artifacts {
			artifacts = append(artifacts, artifact)
		}

		// if the template specifies an "OnlyOn" value, overwrite the current one
		// anyone using "OnlyOn" in a pipeline job needs to understand this isn't an
		// append but an overwrite
		if len(template.OnlyOn) > 0 {
			job.OnlyOn = template.OnlyOn
		}
	}

	// now that we've looped through the templates, we have all of the environment
	// variables and artifacts.

	// if the job had any environment variables specified, add them last (so they
	// overwrite any previously defined variables)
	for _, env := range job.Environment {
		envVars = append(envVars, env)
	}

	// if the job had any artifacts specified, add them last
	for _, artifact := range job.Artifacts {
		artifacts = append(artifacts, artifact)
	}

	// lastly, set the built environment variables + artifacts
	job.Environment = envVars
	job.Artifacts = artifacts

	return job
}

func (p *PipelineExecutor) getTemplateByName(name string) (*api.PipelineSpecJobTemplate, error) {
	name = strings.ToLower(name)

	for _, template := range p._cachedPipeline.Spec.Templates {
		if strings.ToLower(template.Name) == name {
			return &template, nil
		}
	}

	return nil, errors.New("template does not exist")
}
