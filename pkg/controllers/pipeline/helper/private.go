package helper

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (p *PipelineHelper) canRunAnotherPipeline() error {
	return nil
}

func (p *PipelineHelper) processFirstStage() error {
	// mark the pipeline as running
	// create all of the resources we will need to bootstrap the pipeline
	// 		- create minio server (secret, deployment, service)
	//		- create all of the jobs to launch
	// actualize the resources created from the previous step
	// sit back and relax :)

	return nil
}

func (p *PipelineHelper) processRunningPipeline() error {
	return nil
}

func (p *PipelineHelper) getJobsToRun() []api.PipelineSpecJob {
	jobs := []api.PipelineSpecJob{}

	return jobs
}

func (p *PipelineHelper) patchPipeline() error {
	p.pipeline.Status.LastUpdated.Time = time.Now()

	origBytes, err := json.Marshal(p.cachedPipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling original pipeline")
	}

	updatedBytes, err := json.Marshal(p.pipeline)
	if err != nil {
		return errors.Wrap(err, "error marshalling updated pipeline")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return errors.Wrap(err, "error creating json merge patch for pipeline")
	}

	res, err := p.pipelineClient.Pipelines(p.cachedPipeline.Namespace).Patch(p.cachedPipeline.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return errors.Wrap(err, "error patching pipeline")
	}

	p.cachedPipeline = res
	p.pipeline = res.DeepCopy()
	return nil
}
