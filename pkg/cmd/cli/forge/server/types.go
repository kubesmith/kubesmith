package server

import (
	"context"

	"github.com/kubesmith/kubesmith/pkg/controllers"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Namespace string

	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
}

type Server struct {
	options    *Options
	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
	namespace  string

	ctx                   context.Context
	cancelContext         context.CancelFunc
	sharedInformerFactory informers.SharedInformerFactory
	pipelineController    controllers.Interface
}
