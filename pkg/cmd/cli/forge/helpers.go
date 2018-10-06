package forge

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/cli/forge/server"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	c := &cobra.Command{
		Use:   "forge",
		Short: "Orchestrates kubesmith pipelines from a Kubernetes controller",
		Long:  "Orchestrates kubesmith pipelines from a Kubernetes controller",
	}

	c.AddCommand(
		server.NewCommand(f),
	)

	return c
}
