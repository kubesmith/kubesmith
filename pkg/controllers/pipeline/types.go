package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithListersv1 "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
)

type PipelineController struct {
	*generic.GenericController

	logger          logrus.FieldLogger
	kubeClient      kubernetes.Interface
	kubesmithClient kubesmithv1.KubesmithV1Interface

	pipelineLister kubesmithListersv1.PipelineLister
	clock          clock.Clock
}
