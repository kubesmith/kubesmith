package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	"k8s.io/client-go/kubernetes"
)

type PipelineController struct {
	*generic.GenericController

	namespace           string
	maxRunningPipelines int
	kubeClient          kubernetes.Interface
	kubesmithClient     kubesmithv1.KubesmithV1Interface
}
