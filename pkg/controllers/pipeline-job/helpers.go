package pipelinejob

import (
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

func NewPipelineJobController(
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineJobInformer informers.PipelineJobInformer,
) controllers.Interface {
	c := &PipelineJobController{
		GenericController: generic.NewGenericController("PipelineJob"),
		logger:            logger.WithField("controller", "PipelineJob"),
		kubeClient:        kubeClient,
		kubesmithClient:   kubesmithClient,
		pipelineJobLister: pipelineJobInformer.Lister(),
		clock:             &clock.RealClock{},
	}

	c.SyncHandler = c.processPipelineJob
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineJobInformer.Informer().HasSynced,
	)

	pipelineJobInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				job := obj.(*v1.PipelineJob)

				c.Queue.Add(sync.PipelineJobAddAction(*job))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedJob := updatedObj.(*v1.PipelineJob)

				c.Queue.Add(sync.PipelineJobUpdateAction(*updatedJob))
			},
			DeleteFunc: func(obj interface{}) {
				job := obj.(*v1.PipelineJob)

				c.Queue.Add(sync.PipelineJobDeleteAction(*job))
			},
		},
	)

	return c
}
