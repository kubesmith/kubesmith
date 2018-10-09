package helper

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	genApi "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/client-go/kubernetes"
)

func NewPipelineHelper(
	pipeline *api.Pipeline,
	pipelineLister listers.PipelineLister,
	pipelineClient genApi.PipelinesGetter,
	kubeClient kubernetes.Interface,
) *PipelineHelper {
	if pipeline == nil {
		// developer error
		panic("invalid pipeline was passed to the NewPipelineHelper()")
	}

	return &PipelineHelper{
		pipeline:       pipeline.DeepCopy(),
		cachedPipeline: pipeline,

		pipelineLister: pipelineLister,
		pipelineClient: pipelineClient,
		kubeClient:     kubeClient,
	}
}
