package server

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Namespace, "namespace", "", "The namespace where this forge server will run")
	env.BindEnvToFlag("namespace", flags)
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

	if o.Namespace == "" {
		ns, err := getNamespaceFromServiceAccount()
		if err == nil {
			o.Namespace = ns
		} else {
			o.Namespace = api.DefaultNamespace
		}
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
