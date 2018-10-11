package executor

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	v1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func NewPipelineExecutor(
	pipeline api.Pipeline,
	maxRunningPipelines int,
	logger logrus.FieldLogger,
	kubeClient kubernetes.Interface,
	kubesmithClient v1.KubesmithV1Interface,
) *PipelineExecutor {
	return &PipelineExecutor{
		Pipeline:            pipeline.DeepCopy(),
		MaxRunningPipelines: maxRunningPipelines,

		_cachedPipeline: pipeline,

		logger:          logger,
		kubeClient:      kubeClient,
		kubesmithClient: kubesmithClient,
	}
}
