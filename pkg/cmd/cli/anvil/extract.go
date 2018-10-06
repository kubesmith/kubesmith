package anvil

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewExtractCommand(f client.Factory, use string) *cobra.Command {
	o := NewExtractOptions()

	c := &cobra.Command{
		Use:   use,
		Short: "",
		Long:  "",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Complete(args, f))
			cmd.CheckError(o.Validate(c, args, f))
			cmd.CheckError(o.Run(c, f))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

type ExtractOptions struct {
	client kubesmithClient.Interface
}

func NewExtractOptions() *ExtractOptions {
	return &ExtractOptions{
		//
	}
}

func (o *ExtractOptions) BindFlags(flags *pflag.FlagSet) {
	//
}

func (o *ExtractOptions) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *ExtractOptions) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	o.client = client
	return nil
}

func (o *ExtractOptions) Run(c *cobra.Command, f client.Factory) error {
	return nil
}
