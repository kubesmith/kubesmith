package generic

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"k8s.io/client-go/tools/cache"
)

func (c *GenericController) runWorker() {
	for c.processNextWorkItem() {
		//
	}
}

func (c *GenericController) processNextWorkItem() bool {
	key, quit := c.Queue.Get()
	if quit {
		return false
	}
	// always call done on this item, since if it fails we'll add
	// it back with rate-limiting below
	defer c.Queue.Done(key)

	err := c.SyncHandler(key.(sync.SyncAction))
	if err == nil {
		// If you had no error, tell the queue to stop tracking history for your key. This will reset
		// things like failure counts for per-item rate limiting.
		c.Queue.Forget(key)
		return true
	}

	glog.Error("Error in syncHandler, re-adding item to queue")
	// we had an error processing the item so add it back
	// into the queue for re-processing with rate-limiting
	c.Queue.AddRateLimited(key)

	return true
}

func (c *GenericController) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Info("Error creating queue key, item not added to queue")
		glog.Error(err)
		return
	}

	c.Queue.Add(key)
}

func (c *GenericController) enqueueSecond(_, obj interface{}) {
	c.enqueue(obj)
}
