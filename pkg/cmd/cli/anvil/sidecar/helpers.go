package sidecar

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	o := NewOptions()

	c := &cobra.Command{
		Use:   "sidecar",
		Short: "Runs as a sidecar (for Kubernetes) listening for pod completion",
		Long:  "Runs as a sidecar (for Kubernetes) listening for pod completion",
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

func getNamespaceFromServiceAccount() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", errors.New("cannot retrieve namespace from service account")
	}

	return strings.TrimSpace(string(data)), nil
}
