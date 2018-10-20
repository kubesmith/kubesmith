package forge

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

func NewForgeController(
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	forgeInformer informers.ForgeInformer,
) controllers.Interface {
	c := &ForgeController{
		GenericController: generic.NewGenericController("Forge"),
		logger:            logger.WithField("controller", "Forge"),
		kubeClient:        kubeClient,
		kubesmithClient:   kubesmithClient,
		forgeLister:       forgeInformer.Lister(),
		clock:             &clock.RealClock{},
	}

	c.SyncHandler = c.processForge
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		forgeInformer.Informer().HasSynced,
	)

	forgeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				forge := obj.(*v1.Forge)

				c.Queue.Add(sync.ForgeAddAction(*forge))
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				updatedForge := updatedObj.(*v1.Forge)

				c.Queue.Add(sync.ForgeUpdateAction(*updatedForge))
			},
			DeleteFunc: func(obj interface{}) {
				forge := obj.(*v1.Forge)

				c.Queue.Add(sync.ForgeDeleteAction(*forge))
			},
		},
	)

	return c
}
