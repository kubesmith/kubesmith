package forge

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	api "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/tools/cache"
)

func NewForgeController(forgeClient api.ForgesGetter, forgeInformer informers.ForgeInformer) controllers.Interface {
	c := &ForgeController{
		GenericController: generic.NewGenericController("forge"),
		forgeLister:       forgeInformer.Lister(),
		forgeClient:       forgeClient,
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
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					forge := obj.(*v1.Forge)
					glog.Errorf("Error creating queue key, item not added to queue; name: %s", forge.Name)
					return
				}

				c.Queue.Add(key)
			},
		},
	)

	return c
}
