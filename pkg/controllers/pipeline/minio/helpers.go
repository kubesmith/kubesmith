package minio

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	appListersv1 "k8s.io/client-go/listers/apps/v1"
	coreListersv1 "k8s.io/client-go/listers/core/v1"
)

func NewMinioServer(
	namespace, resourcePrefix string,
	logger logrus.FieldLogger,
	resourceLabels map[string]string,
	kubeClient kubernetes.Interface,
	secretLister coreListersv1.SecretLister,
	deploymentLister appListersv1.DeploymentLister,
	serviceLister coreListersv1.ServiceLister,
) *MinioServer {
	return &MinioServer{
		namespace:        namespace,
		resourcePrefix:   resourcePrefix,
		resourceLabels:   resourceLabels,
		logger:           logger,
		kubeClient:       kubeClient,
		secretLister:     secretLister,
		deploymentLister: deploymentLister,
		serviceLister:    serviceLister,
	}
}
