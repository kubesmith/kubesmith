package pipelinestage

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineStageController(
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineInformer informers.PipelineInformer,
	pipelineStageInformer informers.PipelineStageInformer,
	pipelineJobInformer informers.PipelineJobInformer,
) controllers.Interface {
	c := &PipelineStageController{
		GenericController:   generic.NewGenericController("PipelineStage"),
		logger:              logger.WithField("controller", "PipelineStage"),
		kubeClient:          kubeClient,
		kubesmithClient:     kubesmithClient,
		pipelineLister:      pipelineInformer.Lister(),
		pipelineStageLister: pipelineStageInformer.Lister(),
		pipelineJobLister:   pipelineJobInformer.Lister(),
		clock:               &clock.RealClock{},
	}

	c.SyncHandler = c.processPipelineStage
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineInformer.Informer().HasSynced,
		pipelineStageInformer.Informer().HasSynced,
		pipelineJobInformer.Informer().HasSynced,
	)

	pipelineStageInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				stage := obj.(*v1.PipelineStage)

				c.Queue.Add(sync.PipelineStageAddAction(*stage))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedStage := updatedObj.(*v1.PipelineStage)

				c.Queue.Add(sync.PipelineStageUpdateAction(*updatedStage))
			},
			DeleteFunc: func(obj interface{}) {
				switch obj.(type) {
				case cache.DeletedFinalStateUnknown:
					stage := obj.(cache.DeletedFinalStateUnknown).Obj.(*v1.PipelineStage)
					c.Queue.Add(sync.PipelineStageDeleteAction(*stage))
				case *v1.PipelineStage:
					stage := obj.(*v1.PipelineStage)
					c.Queue.Add(sync.PipelineStageDeleteAction(*stage))
				default:
					c.logger.Info("ignoring deleted object; unknown")
					spew.Dump(obj)
				}
			},
		},
	)

	return c
}
