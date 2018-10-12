package minio

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func NewMinioServer(
	namespace, resourcePrefix string,
	resourceLabels map[string]string,
	logger logrus.FieldLogger,
	kubeClient kubernetes.Interface,
) *MinioServer {
	return &MinioServer{
		Namespace:      namespace,
		ResourcePrefix: resourcePrefix,
		ResourceLabels: resourceLabels,

		logger:     logger,
		kubeClient: kubeClient,
	}
}
