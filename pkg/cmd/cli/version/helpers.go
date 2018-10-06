package version

import (
	"fmt"

	"github.com/kubesmith/kubesmith/pkg/buildinfo"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {

	c := &cobra.Command{
		Use:   "version",
		Short: "Prints the version information",
		Long:  "Prints the version information",
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", buildinfo.Version)
			fmt.Println("Git:")
			fmt.Printf("  SHA: %s\n", buildinfo.GitSHA)
		},
	}

	return c
}
