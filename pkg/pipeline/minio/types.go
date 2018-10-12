package minio

import (
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
)

type MinioServer struct {
	Namespace      string
	ResourcePrefix string
	ResourceLabels map[string]string

	logger     logrus.FieldLogger
	kubeClient kubernetes.Interface

	minioSecret     *corev1.Secret
	minioDeployment *appsv1.Deployment
	minioService    *corev1.Service

	deploymentLister appListersv1.DeploymentLister
}
