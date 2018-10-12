package jobs

import (
	"fmt"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/pipeline/minio"
	batchListersv1 "k8s.io/client-go/listers/batch/v1"
)

func GetResourceName(prefix string, stageIndex, jobIndex int) string {
	return fmt.Sprintf("%s-stage-%d-job-%d", prefix, stageIndex, jobIndex)
}

func ScheduleJob(
	name, pipelineName string,
	labels map[string]string,
	job api.PipelineSpecJob,
	minioServer *minio.MinioServer,
	jobLister batchListersv1.JobLister,
) error {
	// todo: left off here
	// be sure to query for the existing jobs to see if they are running (with configmaps, too)
	fmt.Println("todo: write schedule job function")

	return nil
}
