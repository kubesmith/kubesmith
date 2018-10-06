package server

import (
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
}
