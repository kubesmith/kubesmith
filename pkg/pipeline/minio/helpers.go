package minio

import (
	"k8s.io/client-go/kubernetes"
)

func NewMinioServer(
	namespace, resourcePrefix string,
	resourceLabels map[string]string,
	kubeClient kubernetes.Interface,
) *MinioServer {
	return &MinioServer{
		Namespace:      namespace,
		ResourcePrefix: resourcePrefix,
		ResourceLabels: resourceLabels,

		kubeClient: kubeClient,
	}
}
