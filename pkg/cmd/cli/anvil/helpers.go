package anvil

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/anvil/extract"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/anvil/sidecar"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/anvil/wait"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	c := &cobra.Command{
		Use:   "anvil",
		Short: "Manages artifact syncing and job lifecycle",
		Long:  "Manages artifact syncing and job lifecycle",
	}

	c.AddCommand(
		extract.NewCommand(f),
		wait.NewCommand(f),
		sidecar.NewCommand(f),
	)

	return c
}
