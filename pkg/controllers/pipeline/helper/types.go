package helper

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type PipelineHelper struct {
	pipeline       *api.Pipeline
	cachedPipeline *api.Pipeline

	resourcePrefix string
	resourceLabels map[string]string

	minioSecret     *corev1.Secret
	minioDeployment *appsv1.Deployment
	minioService    *corev1.Service

	kubeClient      kubernetes.Interface
	kubesmithClient kubesmithv1.KubesmithV1Interface
}
