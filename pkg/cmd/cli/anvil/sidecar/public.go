package sidecar

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Pod.Name, "pod-name", "", "The name of the pod to monitor")
	env.BindEnvToFlag("pod-name", flags)
	flags.StringVar(&o.Pod.Namespace, "pod-namespace", "", "The namespace of the pod to monitor")
	env.BindEnvToFlag("pod-namespace", flags)
	flags.StringVar(&o.Sidecar.Name, "name", "", "The name of the sidecar container")
	env.BindEnvToFlag("name", flags)
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}
	o.client = client

	kubeClient, err := f.KubeClient()
	if err != nil {
		return err
	}
	o.kubeClient = kubeClient

	if o.Pod.Namespace == "" {
		ns, err := getNamespaceFromServiceAccount()
		if err == nil {
			o.Pod.Namespace = ns
		} else {
			o.Pod.Namespace = "default"
		}
	}

	if o.Sidecar.Name == "" {
		o.Sidecar.Name = "kubesmith"
	}

	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	server := NewServer(o)

	if err := server.run(); err != nil {
		return err
	}

	return nil
}
