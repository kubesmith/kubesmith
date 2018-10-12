package minio

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
)

func NewMinioServer(
	namespace, resourcePrefix string,
	resourceLabels map[string]string,
	logger logrus.FieldLogger,
	kubeClient kubernetes.Interface,
	deploymentLister appListersv1.DeploymentLister,
) *MinioServer {
	return &MinioServer{
		Namespace:        namespace,
		ResourcePrefix:   resourcePrefix,
		ResourceLabels:   resourceLabels,
		logger:           logger,
		kubeClient:       kubeClient,
		deploymentLister: deploymentLister,
	}
}
