package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	api "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

type PipelineController struct {
	*generic.GenericController

	pipelineLister listers.PipelineLister
	pipelineClient api.PipelinesGetter
	clock          clock.Clock
}
