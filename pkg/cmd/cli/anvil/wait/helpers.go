package wait

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	o := NewOptions()

	c := &cobra.Command{
		Use:   "wait",
		Short: "Waits for a flag file to exist and optionally uploads any detected artifacts to a remote s3 bucket",
		Long:  "Waits for a flag file to exist and optionally uploads any detected artifacts to a remote s3 bucket",
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
	return &Options{
		S3: OptionsS3{},
	}
}
