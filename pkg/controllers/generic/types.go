package generic

import (
	"time"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type GenericController struct {
	Name             string
	Queue            workqueue.RateLimitingInterface
	SyncHandler      func(key string) error
	ResyncFunc       func()
	ResyncPeriod     time.Duration
	CacheSyncWaiters []cache.InformerSynced
}
