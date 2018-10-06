package anvil

import (
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	c := &cobra.Command{
		Use:   "anvil",
		Short: "",
		Long:  "",
	}

	c.AddCommand(
		NewExtractCommand(f, "extract"),
		NewWaitCommand(f, "wait"),
	)

	return c
}
