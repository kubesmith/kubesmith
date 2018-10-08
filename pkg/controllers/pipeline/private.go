package pipeline

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/minio"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

func (c *PipelineController) processPipeline(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	pipeline, err := c.pipelineLister.Pipelines(ns).Get(name)
	if apierrors.IsNotFound(err) {
		glog.V(1).Info("unable to find pipeline")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting pipeline")
	}

	if !c.pipelineHasWork(pipeline) {
		glog.V(1).Info("pipeline had work when added to cache queue but not anymore")
		return nil
	}

	switch pipeline.Status.Phase {
	case api.PipelinePhaseEmpty, api.PipelinePhaseQueued:
		if err := c.processFirstStage(pipeline); err != nil {
			return err
		}
	case api.PipelinePhaseRunning:
		if err := c.processRunningPipeline(pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (c *PipelineController) pipelineHasWork(pipeline *api.Pipeline) bool {
	return true
}

func (c *PipelineController) validatePipeline(pipeline *api.Pipeline) error {
	return nil
}

func (c *PipelineController) canRunAnotherPipeline() (bool, error) {
	pipelines, err := c.pipelineLister.List(labels.Everything())
	if err != nil {
		glog.V(1).Info("could not list pipelines")
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines {
		if pipeline.Status.Phase == api.PipelinePhaseRunning {
			currentlyRunning++
		}
	}

	if currentlyRunning < c.maxRunningPipelines {
		return true, nil
	}

	return false, nil
}

func (c *PipelineController) startStageJobs(pipeline *api.Pipeline) error {
	return nil
}

func (c *PipelineController) markPipelineAsRunning(pipeline *api.Pipeline) (*api.Pipeline, error) {
	updated := pipeline.DeepCopy()
	updated.Status.Phase = api.PipelinePhaseRunning
	updated.Status.StageIndex = 1
	updated.Status.LastUpdated.Time = time.Now()

	return c.patchPipeline(pipeline, updated)
}

func (c *PipelineController) markPipelineAsQueued(pipeline *api.Pipeline) (*api.Pipeline, error) {
	updated := pipeline.DeepCopy()
	updated.Status.Phase = api.PipelinePhaseQueued
	updated.Status.StageIndex = 0
	updated.Status.LastUpdated.Time = time.Now()

	return c.patchPipeline(pipeline, updated)
}

func (c *PipelineController) processFirstStage(pipeline *api.Pipeline) error {
	glog.V(1).Info("validating pipeline")
	if err := c.validatePipeline(pipeline); err != nil {
		return errors.Wrap(err, "could not validate pipeline")
	}

	glog.V(1).Info("checking to see if another pipeline can be run")
	runnable, err := c.canRunAnotherPipeline()
	if !runnable {
		glog.V(1).Info("another pipeline cannot be run")

		if _, err := c.markPipelineAsQueued(pipeline); err != nil {
			glog.V(1).Info("could not set pipeline to queued")
			return errors.Wrap(err, "could not mark pipeline as queued")
		}

		return errors.Wrap(err, "cannot run another pipeline at the moment")
	} else if err != nil {
		glog.V(1).Info("could not check if we can can run another pipeline")
		return errors.Wrap(err, "could not check if we can run another pipeline")
	}

	glog.V(1).Info("setting pipeline to running and stage 1")
	if pipeline, err = c.markPipelineAsRunning(pipeline); err != nil {
		glog.V(1).Info("could not mark pipeline as running")
		return errors.Wrap(err, "could not mark pipeline as running")
	}

	// provision the minio server for this pipeline (for workspace storage)
	glog.V(1).Info("provisioning minio server")
	if err := minio.CreateMinioServerForPipeline(pipeline, c.kubeClient); err != nil {
		glog.V(1).Info("could not provision minio server")
		glog.Error(err)

		if _, err := c.markPipelineAsQueued(pipeline); err != nil {
			glog.V(1).Info("could not set the pipeline to queued")
			return errors.Wrap(err, "could not mark the pipeline as queued")
		}

		return errors.Wrap(err, "could not provision minio server")
	}

	// start all of this stage's pods
	glog.V(1).Info("starting stage jobs")
	if err := c.startStageJobs(pipeline); err != nil {
		glog.V(1).Info("could not start stage jobs; unprovisioning minio server")
		if err := minio.DeleteMinioServerForPipeline(pipeline, c.kubeClient); err != nil {
			glog.V(1).Info("could not unprovision minio server")
			return errors.Wrap(err, "could not unprovision minio server")
		}

		glog.V(1).Info("setting pipeline back to queued")
		if _, err := c.markPipelineAsQueued(pipeline); err != nil {
			glog.V(1).Info("could not set pipeline back to queued")
			return errors.Wrap(err, "could not patch the pipeline")
		}

		return errors.Wrap(err, "could not start stage jobs")
	}

	glog.V(1).Info("started stage jobs")
	return nil
}

func (c *PipelineController) processRunningPipeline(pipeline *api.Pipeline) error {
	// shift copies so we can continue to do things
	original := pipeline
	pipeline = original.DeepCopy()

	// todo: figure this out
	return nil
}

func (c *PipelineController) resync() {
	list, err := c.pipelineLister.List(labels.Everything())
	if err != nil {
		glog.V(1).Info("error listing pipelines")
		glog.Error(err)
		return
	}

	for _, forge := range list {
		key, err := cache.MetaNamespaceKeyFunc(forge)
		if err != nil {
			glog.Errorf("error generating key for pipeline; key: %s", forge.Name)
			continue
		}

		c.Queue.Add(key)
	}
}

func (c *PipelineController) patchPipeline(original, updated *api.Pipeline) (*api.Pipeline, error) {
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

	res, err := c.pipelineClient.Pipelines(original.Namespace).Patch(original.Name, types.MergePatchType, patchBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error patching pipeline")
	}

	return res, nil
}
