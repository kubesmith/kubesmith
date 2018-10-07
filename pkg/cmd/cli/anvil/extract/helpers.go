package extract

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	o := NewOptions()

	c := &cobra.Command{
		Use:   "extract",
		Short: "Extracts remote s3 archive(s) locally",
		Long:  "Extracts remote s3 archive(s) locally",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Complete(args, f))
			cmd.CheckError(o.Validate(c, args, f))
			cmd.CheckError(o.Run(c, f))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

func NewOptions() *Options {
	return &Options{}
}
