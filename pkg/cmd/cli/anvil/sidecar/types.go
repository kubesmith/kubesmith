package sidecar

import (
	"context"

	"github.com/kubesmith/kubesmith/pkg/controllers"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/sirupsen/logrus"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Sidecar OptionsSidecar
	Pod     OptionsPod

	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
}

type OptionsPod struct {
	Name      string
	Namespace string
}

type OptionsSidecar struct {
	Name string
}

type Server struct {
	options    *Options
	logger     *logrus.Logger
	kubeClient kubernetes.Interface
	namespace  string

	ctx                 context.Context
	cancelContext       context.CancelFunc
	kubeInformerFactory kubeInformers.SharedInformerFactory

	anvilSideCarController controllers.Interface
}
