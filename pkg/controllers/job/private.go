package job

import (
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *JobController) processJob(action sync.SyncAction) error {
	job := action.GetObject().(*batchv1.Job)
	if job == nil {
		c.logger.Panic(errors.New("programmer error; job is nil"))
	}

	job, err := c.jobLister.Jobs(job.GetNamespace()).Get(job.GetName())
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.logger.Infof("skipping: %s/%s", job.GetName(), job.GetNamespace())
			return nil
		}

		return errors.Wrap(err, "error getting job")
	}

	return nil
}
