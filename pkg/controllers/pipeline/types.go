package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	api "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
)

type PipelineController struct {
	*generic.GenericController

	namespace           string
	maxRunningPipelines int

	pipelineLister listers.PipelineLister
	pipelineClient api.PipelinesGetter
	kubeClient     kubernetes.Interface
	clock          clock.Clock
}
