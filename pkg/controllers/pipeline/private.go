package pipeline

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/pipeline/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// create a new logger for this pipeline's execution
	fieldLogger := c.logger.WithFields(logrus.Fields{
		"Name":       pipeline.Name,
		"Namespace":  pipeline.Namespace,
		"Phase":      pipeline.Status.Phase,
		"StageIndex": pipeline.Status.StageIndex,
	})

	// create a new pipeline executor to handle carrying the actions necessary for
	// pipeline to run
	pipelineExecutor := executor.NewPipelineExecutor(
		*pipeline,
		c.maxRunningPipelines,
		fieldLogger,
		c.kubeClient,
		c.kubesmithClient,
		c.pipelineLister,
		c.deploymentLister,
		c.jobLister,
		c.configMapLister,
	)

	// finally, let's execute the pipeline
	fieldLogger.Info("running pipeline executor...")
	if err := pipelineExecutor.Execute(); err != nil {
		fieldLogger.Info("could not run pipeline executor")
		fieldLogger.Error(err)
		return err
	}
	fieldLogger.Info("pipeline executor finished")

	return nil
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
