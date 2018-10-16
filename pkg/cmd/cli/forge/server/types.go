package server

import (
	"context"

	"github.com/kubesmith/kubesmith/pkg/controllers"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	kubesmithInformers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions"
	"github.com/sirupsen/logrus"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Namespace           string
	MaxRunningPipelines int

	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
}

type Server struct {
	options    *Options
	client     kubesmithClient.Interface
	logger     *logrus.Logger
	kubeClient kubernetes.Interface
	namespace  string

	ctx                      context.Context
	cancelContext            context.CancelFunc
	kubesmithInformerFactory kubesmithInformers.SharedInformerFactory
	kubeInformerFactory      kubeInformers.SharedInformerFactory
	pipelineController       controllers.Interface
	pipelineJobController    controllers.Interface
}
