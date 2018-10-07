package generic

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

func (c *GenericController) Run(ctx context.Context, numWorkers int) error {
	if c.SyncHandler == nil && c.ResyncFunc == nil {
		panic("at least one of syncHandler or resyncFunc is required")
	}

	var wg sync.WaitGroup

	defer func() {
		glog.V(1).Info("Waiting for workers to finish their work...")

		c.Queue.ShutDown()
		wg.Wait()

		glog.V(1).Info("All workers have finished")
	}()

	glog.V(1).Info("Starting controller")
	defer glog.V(1).Info("Shutting down controller")

	glog.V(1).Info("Waiting for caches to sync")
	if !cache.WaitForCacheSync(ctx.Done(), c.CacheSyncWaiters...) {
		return errors.New("timed out waiting for caches to sync")
	}
	glog.V(1).Info("Caches are synced")

	if c.SyncHandler != nil {
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func() {
				wait.Until(c.runWorker, time.Second, ctx.Done())
				wg.Done()
			}()
		}
	}

	if c.ResyncFunc != nil {
		if c.ResyncPeriod == 0 {
			panic("non-zero resyncPeriod is required")
		}

		wg.Add(1)
		go func() {
			wait.Until(c.ResyncFunc, c.ResyncPeriod, ctx.Done())
			wg.Done()
		}()
	}

	<-ctx.Done()

	return nil
}
