package pipeline

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func (c *PipelineController) pipelineHasWork(pipeline *v1.Pipeline) bool {
	//

	return true
}

func (c *PipelineController) processPipeline(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	pipeline, err := c.pipelineLister.Pipelines(ns).Get(name)
	if err != nil {
		return errors.Wrap(err, "error getting pipeline")
	}

	spew.Dump(pipeline)

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
