package server

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
	//
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

	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	return nil
}
