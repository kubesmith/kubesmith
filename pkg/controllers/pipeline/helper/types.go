package helper

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	genApi "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/client-go/kubernetes"
)

type PipelineHelper struct {
	pipeline       *api.Pipeline
	cachedPipeline *api.Pipeline

	pipelineLister listers.PipelineLister
	pipelineClient genApi.PipelinesGetter
	kubeClient     kubernetes.Interface
}
