package anvil

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewWaitCommand(f client.Factory, use string) *cobra.Command {
	o := NewWaitOptions()

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

type WaitOptions struct {
	client kubesmithClient.Interface
}

func NewWaitOptions() *WaitOptions {
	return &WaitOptions{
		//
	}
}

func (o *WaitOptions) BindFlags(flags *pflag.FlagSet) {
	//
}

func (o *WaitOptions) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *WaitOptions) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	o.client = client
	return nil
}

func (o *WaitOptions) Run(c *cobra.Command, f client.Factory) error {
	return nil
}
