package helper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	genApi "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPipelineMD5(pipeline *api.Pipeline) string {
	hasher := md5.New()
	hasher.Write([]byte(pipeline.Name))

	return hex.EncodeToString(hasher.Sum(nil))
}

func GetPipelineResourcePrefix(pipeline *api.Pipeline) string {
	return fmt.Sprintf("pipeline-%s", GetPipelineMD5(pipeline))
}

func GetPipelineResourceLabels(pipeline *api.Pipeline) map[string]string {
	labels := map[string]string{
		"PipelineName": pipeline.Name,
		"PipelineMD5":  GetPipelineMD5(pipeline),
	}

	return labels
}

func NewPipelineHelper(
	pipeline *api.Pipeline,
	pipelineLister listers.PipelineLister,
	pipelineClient genApi.PipelinesGetter,
	kubeClient kubernetes.Interface,
) *PipelineHelper {
	if pipeline == nil {
		// developer error
		panic("invalid pipeline helper")
	}

	return &PipelineHelper{
		pipeline:       pipeline.DeepCopy(),
		cachedPipeline: pipeline,

		resourcePrefix: GetPipelineResourcePrefix(pipeline),
		resourceLabels: GetPipelineResourceLabels(pipeline),

		pipelineLister: pipelineLister,
		pipelineClient: pipelineClient,
		kubeClient:     kubeClient,
	}
}