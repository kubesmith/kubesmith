package helper

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (p *PipelineHelper) canRunAnotherPipeline() error {
	return nil
}

func (p *PipelineHelper) processRunningPipeline() error {
	glog.V(1).Info("todo: processing running pipeline...")
	return nil
}

func (p *PipelineHelper) processFinishedPipeline() error {
	// make sure any resources associated to this pipeline's execution are
	// cleaned up

	glog.V(1).Info("todo: processing finished pipeline")
	return nil
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

	res, err := p.kubesmithClient.Pipelines(p.cachedPipeline.Namespace).Patch(p.cachedPipeline.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return errors.Wrap(err, "error patching pipeline")
	}

	p.cachedPipeline = res
	p.pipeline = res.DeepCopy()
	return nil
}
