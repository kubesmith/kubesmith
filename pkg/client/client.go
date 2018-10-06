package client

import (
	"fmt"
	"runtime"

	"github.com/kubesmith/kubesmith/pkg/buildinfo"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config returns a *rest.Config, using either the kubeconfig (if specified) or an in-cluster
// configuration.
func Config(kubeconfig, kubecontext, baseName string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	clientConfig.UserAgent = buildUserAgent(
		baseName,
		buildinfo.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	return clientConfig, nil
}

// buildUserAgent builds a User-Agent string from given args.
func buildUserAgent(command, version, os, arch string) string {
	return fmt.Sprintf("%s/%s (%s/%s) %s", command, version, os, arch)
}
