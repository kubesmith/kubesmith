package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithListersv1 "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
	batchListersv1 "k8s.io/client-go/listers/batch/v1"
)

type PipelineController struct {
	*generic.GenericController

	namespace           string
	maxRunningPipelines int

	logger          *logrus.Logger
	kubeClient      kubernetes.Interface
	kubesmithClient kubesmithv1.KubesmithV1Interface

	pipelineLister   kubesmithListersv1.PipelineLister
	deploymentLister appListersv1.DeploymentLister
	jobLister        batchListersv1.JobLister
}
