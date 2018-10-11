package minio

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type MinioServer struct {
	Namespace      string
	ResourcePrefix string
	ResourceLabels map[string]string

	kubeClient kubernetes.Interface

	minioSecret     *corev1.Secret
	minioDeployment *appsv1.Deployment
	minioService    *corev1.Service
}
