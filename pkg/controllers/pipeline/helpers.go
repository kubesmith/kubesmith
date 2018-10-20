package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
	appInformersv1 "k8s.io/client-go/informers/apps/v1"
	coreInformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineController(
	maxRunningPipelines int,
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineInformer informers.PipelineInformer,
	pipelineStageInformer informers.PipelineStageInformer,
	secretInformer coreInformersv1.SecretInformer,
	deploymentInformer appInformersv1.DeploymentInformer,
	serviceInformer coreInformersv1.ServiceInformer,
) controllers.Interface {
	c := &PipelineController{
		GenericController:   generic.NewGenericController("Pipeline"),
		maxRunningPipelines: maxRunningPipelines,
		logger:              logger.WithField("controller", "Pipeline"),
		kubeClient:          kubeClient,
		kubesmithClient:     kubesmithClient,
		pipelineLister:      pipelineInformer.Lister(),
		pipelineStageLister: pipelineStageInformer.Lister(),
		secretLister:        secretInformer.Lister(),
		deploymentLister:    deploymentInformer.Lister(),
		serviceLister:       serviceInformer.Lister(),
		clock:               &clock.RealClock{},
	}

	c.SyncHandler = c.processPipeline
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineInformer.Informer().HasSynced,
		pipelineStageInformer.Informer().HasSynced,
		secretInformer.Informer().HasSynced,
		deploymentInformer.Informer().HasSynced,
		serviceInformer.Informer().HasSynced,
	)

	pipelineInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pipeline := obj.(*v1.Pipeline)

				c.Queue.Add(sync.PipelineAddAction(pipeline))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedPipeline := updatedObj.(*v1.Pipeline)

				c.Queue.Add(sync.PipelineUpdateAction(updatedPipeline))
			},
			DeleteFunc: func(obj interface{}) {
				pipeline := obj.(*v1.Pipeline)

				c.Queue.Add(sync.PipelineDeleteAction(pipeline))
			},
		},
	)

	return c
}
