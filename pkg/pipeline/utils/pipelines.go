package utils

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

func PatchPipeline(original, updated v1.Pipeline, client kubesmithv1.KubesmithV1Interface) (*v1.Pipeline, error) {
	updated.Status.LastUpdated.Time = time.Now()

	origBytes, err := json.Marshal(original)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling original pipeline")
	}

	updatedBytes, err := json.Marshal(updated)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling updated pipeline")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating json merge patch for pipeline")
	}

	res, err := client.Pipelines(original.Name).Patch(original.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error patching pipeline")
	}

	return res, nil
}

func SetPipelineToQueued(pipeline v1.Pipeline) v1.Pipeline {
	pipeline.Status.StageIndex = 0
	pipeline.Status.Phase = v1.PipelinePhaseQueued

	return pipeline
}

func SetPipelineToCompleted(pipeline v1.Pipeline) v1.Pipeline {
	pipeline.Status.StageIndex = len(pipeline.Spec.Stages)
	pipeline.Status.Phase = v1.PipelinePhaseCompleted
	pipeline.Status.EndTime.Time = time.Now()

	return pipeline
}

func SetPipelineToRunning(pipeline v1.Pipeline) v1.Pipeline {
	pipeline.Status.StageIndex = 1
	pipeline.Status.Phase = v1.PipelinePhaseRunning
	pipeline.Status.StartTime.Time = time.Now()

	return pipeline
}

func SetPipelineToFailed(pipeline v1.Pipeline, reason string) v1.Pipeline {
	pipeline.Status.StageIndex = len(pipeline.Spec.Stages)
	pipeline.Status.Phase = v1.PipelinePhaseFailed
	pipeline.Status.EndTime.Time = time.Now()
	pipeline.Status.FailureReason = reason

	return pipeline
}
