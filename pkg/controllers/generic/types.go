package generic

import (
	"time"

	"github.com/kubesmith/kubesmith/pkg/sync"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type GenericController struct {
	Name             string
	Queue            workqueue.RateLimitingInterface
	SyncHandler      func(action sync.SyncAction) error
	ResyncFunc       func()
	ResyncPeriod     time.Duration
	CacheSyncWaiters []cache.InformerSynced
}
