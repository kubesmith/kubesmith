package anvilsidecar

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	coreInformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewAnvilSidecarController(
	sidecarName string,
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	podInformer coreInformersv1.PodInformer,
) controllers.Interface {
	c := &AnvilSidecarController{
		GenericController: generic.NewGenericController("AnvilSidecar"),
		sidecarName:       sidecarName,
		logger:            logger.WithField("controller", "AnvilSidecar"),
		kubeClient:        kubeClient,
		podLister:         podInformer.Lister(),
		clock:             &clock.RealClock{},
	}

	c.SyncHandler = c.processPod
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		podInformer.Informer().HasSynced,
	)

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)

				c.Queue.Add(sync.PodAddAction(*pod))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedPod := updatedObj.(*corev1.Pod)

				c.Queue.Add(sync.PodUpdateAction(*updatedPod))
			},
			DeleteFunc: func(obj interface{}) {
				switch obj.(type) {
				case cache.DeletedFinalStateUnknown:
					pod := obj.(cache.DeletedFinalStateUnknown).Obj.(*corev1.Pod)
					c.Queue.Add(sync.PodDeleteAction(*pod))
				case *corev1.Pod:
					pod := obj.(*corev1.Pod)
					c.Queue.Add(sync.PodDeleteAction(*pod))
				default:
					c.logger.Info("ignoring deleted object; unknown")
					spew.Dump(obj)
				}
			},
		},
	)

	return c
}
