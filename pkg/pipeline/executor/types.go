package executor

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	v1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithListersv1 "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
	batchListersv1 "k8s.io/client-go/listers/batch/v1"
	coreListersv1 "k8s.io/client-go/listers/core/v1"
)

type PipelineExecutor struct {
	MaxRunningPipelines int

	_cachedPipeline api.Pipeline

	logger          logrus.FieldLogger
	kubeClient      kubernetes.Interface
	kubesmithClient v1.KubesmithV1Interface

	pipelineLister   kubesmithListersv1.PipelineLister
	deploymentLister appListersv1.DeploymentLister
	jobLister        batchListersv1.JobLister
	configMapLister  coreListersv1.ConfigMapLister
}
