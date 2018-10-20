package job

import (
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	batchInformersv1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewJobController(
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	jobInformer batchInformersv1.JobInformer,
) controllers.Interface {
	c := &JobController{
		GenericController: generic.NewGenericController("Job"),
		logger:            logger.WithField("controller", "Job"),
		kubeClient:        kubeClient,
		kubesmithClient:   kubesmithClient,
		jobLister:         jobInformer.Lister(),
		clock:             &clock.RealClock{},
	}

	c.SyncHandler = c.processJob
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		jobInformer.Informer().HasSynced,
	)

	jobInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				job := obj.(*batchv1.Job)

				c.Queue.Add(sync.JobAddAction(job))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedJob := updatedObj.(*batchv1.Job)

				c.Queue.Add(sync.JobUpdateAction(updatedJob))
			},
			DeleteFunc: func(obj interface{}) {
				job := obj.(*batchv1.Job)

				c.Queue.Add(sync.JobDeleteAction(job))
			},
		},
	)

	return c
}
