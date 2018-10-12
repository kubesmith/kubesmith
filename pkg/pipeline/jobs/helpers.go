package jobs

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/minio"
)

func EnsureJobsAreScheduled(jobs []api.PipelineSpecJob, minioServer *minio.MinioServer) error {
	// todo: left off here
	// be sure to query for the existing jobs to see if they are running (with configmaps, too)
	return nil
}
