package pipeline

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	api "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineController(pipelineClient api.PipelinesGetter, pipelineInformer informers.PipelineInformer) controllers.Interface {
	c := &PipelineController{
		GenericController: generic.NewGenericController("pipeline"),
		pipelineLister:    pipelineInformer.Lister(),
		pipelineClient:    pipelineClient,
		clock:             &clock.RealClock{},
	}

	c.SyncHandler = c.processPipeline
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineInformer.Informer().HasSynced,
	)

	pipelineInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				// TODO: FILTER PIPELINES TO ONLY RELEVANT ONES (NO RUNNING OR FINISHED)
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					pipeline := obj.(*v1.Pipeline)
					glog.Errorf("Error creating queue key, item not added to queue; name: %s", pipeline.Name)
					return
				}

				c.Queue.Add(key)
			},

			// HAVE NOT DECIDED IF I WANT UPDATE/DELETE REACTIONS IN THIS CONTROLLER YET

			// UpdateFunc: func(old, new interface{}) {
			// 	key, err := cache.MetaNamespaceKeyFunc(new)
			// 	if err != nil {
			// 		newPipeline := new.(*v1.Pipeline)
			// 		glog.Errorf("Error updating queue key, item not added to queue; name: %s", newPipeline.Name)
			// 		return
			// 	}

			// 	c.Queue.Add(key)
			// },
			// DeleteFunc: func(obj interface{}) {
			// 	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			// 	if err != nil {
			// 		pipeline := obj.(*v1.Pipeline)
			// 		glog.Errorf("Error deleting queue key, item not added to queue; name: %s", pipeline.Name)
			// 		return
			// 	}

			// 	c.Queue.Add(key)
			// },
		},
	)

	return c
}
