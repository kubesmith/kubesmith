package job

import (
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
)

func (c *JobController) processJob(action sync.SyncAction) error {
	cachedJob := action.GetObject().(batchv1.Job)
	job, err := c.jobLister.Jobs(cachedJob.GetNamespace()).Get(cachedJob.GetName())
	if err != nil {
		return errors.Wrap(err, "error getting job")
	}

	_ = job
	return nil
}
