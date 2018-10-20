package minio

import (
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
	coreListersv1 "k8s.io/client-go/listers/core/v1"
)

const (
	MINIO_PORT = 9000
)

type MinioServer struct {
	namespace      string
	resourcePrefix string
	resourceLabels map[string]string

	logger           logrus.FieldLogger
	kubeClient       kubernetes.Interface
	secretLister     coreListersv1.SecretLister
	deploymentLister appListersv1.DeploymentLister
	serviceLister    coreListersv1.ServiceLister

	minioSecret     *corev1.Secret
	minioDeployment *appsv1.Deployment
	minioService    *corev1.Service
}
