package pipelinejob

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
	batchInformersv1 "k8s.io/client-go/informers/batch/v1"
	coreInformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineJobController(
	maxRunningPipelineJobs int,
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineJobInformer informers.PipelineJobInformer,
	pipelineStageInformer informers.PipelineStageInformer,
	configMapInformer coreInformersv1.ConfigMapInformer,
	jobInformer batchInformersv1.JobInformer,
) controllers.Interface {
	c := &PipelineJobController{
		GenericController:      generic.NewGenericController("PipelineJob"),
		maxRunningPipelineJobs: maxRunningPipelineJobs,
		logger:                 logger.WithField("controller", "PipelineJob"),
		kubeClient:             kubeClient,
		kubesmithClient:        kubesmithClient,
		pipelineJobLister:      pipelineJobInformer.Lister(),
		pipelineStageLister:    pipelineStageInformer.Lister(),
		configMapLister:        configMapInformer.Lister(),
		jobLister:              jobInformer.Lister(),
		clock:                  &clock.RealClock{},
	}

	c.SyncHandler = c.processPipelineJob
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineJobInformer.Informer().HasSynced,
		pipelineStageInformer.Informer().HasSynced,
		configMapInformer.Informer().HasSynced,
		jobInformer.Informer().HasSynced,
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
				switch obj.(type) {
				case cache.DeletedFinalStateUnknown:
					pipelineJob := obj.(cache.DeletedFinalStateUnknown).Obj.(*v1.PipelineJob)
					c.Queue.Add(sync.PipelineJobDeleteAction(*pipelineJob))
				case *v1.PipelineJob:
					pipelineJob := obj.(*v1.PipelineJob)
					c.Queue.Add(sync.PipelineJobDeleteAction(*pipelineJob))
				default:
					c.logger.Info("ignoring deleted object; unknown")
					spew.Dump(obj)
				}
			},
		},
	)

	return c
}
