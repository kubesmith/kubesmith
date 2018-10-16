package pipelinejob

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithInformers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
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
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		jobsInformer.Informer().HasSynced,
	)

	// setup event handlers for jobs
	jobsInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				job := obj.(*batchv1.Job)

				key, err := cache.MetaNamespaceKeyFunc(job)
				if err != nil {
					glog.Errorf("Error creating queue key, item not added to queue; name: %s", job.Name)
					return
				}

				// make sure this feels like a valid pipeline job
				if isPipelineJob := c.jobIsPipelineJob(job); !isPipelineJob {
					return
				}

				// everything existed, add this job to the queue to be processed
				c.Queue.Add(key)
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedJob := updatedObj.(*batchv1.Job)

				key, err := cache.MetaNamespaceKeyFunc(updatedJob)
				if err != nil {
					glog.Errorf("Error updating queue key, item not added to queue; name: %s", updatedJob.Name)
					return
				}

				// make sure this feels like a valid pipeline job
				if isPipelineJob := c.jobIsPipelineJob(updatedJob); !isPipelineJob {
					return
				}

				// everything existed, add this job to the queue to be processed
				c.Queue.Add(key)
			},
		},
	)

	return c
}
