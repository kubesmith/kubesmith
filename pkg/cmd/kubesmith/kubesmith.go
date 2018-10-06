package kubesmith

import (
	"flag"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/anvil"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/forge"
	"github.com/spf13/cobra"
)

func NewCommand(name string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: "",
		Long:  "",
	}

	f := client.NewFactory(name)
	f.BindFlags(c.PersistentFlags())

	c.AddCommand(
		anvil.NewCommand(f),
		forge.NewCommand(f),
	)

	// add the glog flags
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// TODO: switch to a different logging library.
	// Work around https://github.com/golang/glog/pull/13.
	// See also https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	return c
}
