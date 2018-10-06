package forge

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewCommand(f client.Factory) *cobra.Command {
	o := NewOptions()

	c := &cobra.Command{
		Use:   "forge",
		Short: "Orchestrates kubesmith pipelines from a Kubernetes controller",
		Long:  "Orchestrates kubesmith pipelines from a Kubernetes controller",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Complete(args, f))
			cmd.CheckError(o.Validate(c, args, f))
			cmd.CheckError(o.Run(c, f))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

type Options struct {
	client kubesmithClient.Interface
}

func NewOptions() *Options {
	return &Options{
		//
	}
}

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
	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	return nil
}
