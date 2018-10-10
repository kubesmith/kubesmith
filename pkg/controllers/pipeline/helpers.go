package pipeline

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineController(
	namespace string,
	maxRunningPipelines int,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineInformer informers.PipelineInformer,
) controllers.Interface {
	c := &PipelineController{
		GenericController:   generic.NewGenericController("pipeline"),
		namespace:           namespace,
		maxRunningPipelines: maxRunningPipelines,
		kubeClient:          kubeClient,
		kubesmithClient:     kubesmithClient,
	}

	c.SyncHandler = c.processPipeline
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineInformer.Informer().HasSynced,
	)

	pipelineInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pipeline := obj.(*v1.Pipeline)

				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Errorf("Error creating queue key, item not added to queue; name: %s", pipeline.Name)
					return
				}

				c.Queue.Add(key)
			},
		},
	)

	return c
}
