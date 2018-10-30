package anvilsidecar

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
	coreListersv1 "k8s.io/client-go/listers/core/v1"
)

type AnvilSidecarController struct {
	*generic.GenericController

	sidecarName string
	logger      logrus.FieldLogger
	kubeClient  kubernetes.Interface

	podLister coreListersv1.PodLister
	clock     clock.Clock
}
