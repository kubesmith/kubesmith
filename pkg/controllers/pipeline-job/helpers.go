package pipelinejob

import (
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithInformers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	batchInformersv1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineJobController(
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineInformer kubesmithInformers.PipelineInformer,
	jobsInformer batchInformersv1.JobInformer,
) controllers.Interface {
	c := &PipelineJobController{
		GenericController: generic.NewGenericController("PipelineJob"),
		logger:            logger.WithField("controller", "PipelineJob"),
		kubeClient:        kubeClient,
		kubesmithClient:   kubesmithClient,
		pipelineLister:    pipelineInformer.Lister(),
		jobLister:         jobsInformer.Lister(),
	}

	c.SyncHandler = c.processPipelineJob
	c.ResyncFunc = c.resyncJobs
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		jobsInformer.Informer().HasSynced,
	)

	// setup event handlers for jobs
	jobsInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				job := obj.(*batchv1.Job)

				if isPipelineJob := c.jobIsPipelineJob(job); !isPipelineJob {
					return
				}

				c.Queue.Add(sync.JobAddAction(job))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedJob := updatedObj.(*batchv1.Job)

				if isPipelineJob := c.jobIsPipelineJob(updatedJob); !isPipelineJob {
					return
				}

				c.Queue.Add(sync.JobUpdateAction(updatedJob))
			},
		},
	)

	return c
}
