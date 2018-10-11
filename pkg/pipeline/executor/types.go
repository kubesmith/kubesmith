package executor

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	v1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type PipelineExecutor struct {
	Pipeline            *api.Pipeline
	MaxRunningPipelines int

	_cachedPipeline api.Pipeline

	logger          logrus.FieldLogger
	kubeClient      kubernetes.Interface
	kubesmithClient v1.KubesmithV1Interface
}
