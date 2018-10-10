package pipeline

import (
	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline/helper"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (c *PipelineController) processPipeline(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	pipeline, err := c.kubesmithClient.Pipelines(ns).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		glog.V(1).Info("unable to find pipeline")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting pipeline")
	}

	// create a new pipeline helper that can assist with making things easier
	pipelineHelper := helper.NewPipelineHelper(
		pipeline,
		c.kubeClient,
		c.kubesmithClient,
	)

	// check to see if the forge can run another pipeline in this namespace
	glog.V(1).Info("checking to see if another pipeline can be run")
	runnable, err := c.canRunAnotherPipeline()
	if !runnable {
		glog.V(1).Info("another pipeline cannot be run")

		if err := pipelineHelper.SetPipelineStatus(api.PipelinePhaseQueued); err != nil {
			glog.V(1).Info("could not set pipeline to queued")
			return err
		}

		return err
	} else if err != nil {
		glog.V(1).Info("could not check if we can can run another pipeline")
		return err
	}

	glog.V(1).Info("executing pipeline")
	if err := pipelineHelper.Execute(); err != nil {
		glog.V(1).Info("could not execute pipeline")
		glog.Error(err)
		return err
	}
	glog.V(1).Info("pipeline executed")

	return nil
}

func (c *PipelineController) canRunAnotherPipeline() (bool, error) {
	pipelines, err := c.kubesmithClient.Pipelines(c.namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.V(1).Info("could not list pipelines")
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines.Items {
		if pipeline.Status.Phase == api.PipelinePhaseRunning {
			currentlyRunning++
		}
	}

	if currentlyRunning < c.maxRunningPipelines {
		return true, nil
	}

	return false, nil
}

func (c *PipelineController) resync() {
	list, err := c.kubesmithClient.Pipelines(c.namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.V(1).Info("error listing pipelines")
		glog.Error(err)
		return
	}

	for _, forge := range list.Items {
		key, err := cache.MetaNamespaceKeyFunc(forge)
		if err != nil {
			glog.Errorf("error generating key for pipeline; key: %s", forge.Name)
			continue
		}

		c.Queue.Add(key)
	}
}
