package server

import kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"

type Options struct {
	client kubesmithClient.Interface
}
