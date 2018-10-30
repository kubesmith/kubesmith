package sidecar

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	anvilsidecar "github.com/kubesmith/kubesmith/pkg/controllers/anvil-sidecar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeInformers "k8s.io/client-go/informers"
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

func NewServer(o *Options) *Server {
	ctx, cancelContext := context.WithCancel(context.Background())

	// call the cancelContext function if we receive a interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		glog.Infof("Received signal %s, shutting down", sig)
		cancelContext()
	}()

	// setup a logger for this pipeline controller
	logger := logrus.New()

	// setup our informers
	kubeInformerFactory := kubeInformers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		0,
		kubeInformers.WithNamespace(o.Pod.Namespace),
		kubeInformers.WithTweakListOptions(func(listOptions *metav1.ListOptions) {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", o.Pod.Name)
		}),
	)

	// setup our controllers
	anvilSideCarController := anvilsidecar.NewAnvilSidecarController(
		o.Sidecar.Name,
		logger,
		o.kubeClient,
		kubeInformerFactory.Core().V1().Pods(),
	)

	// finally, return the server
	return &Server{
		options:                o,
		logger:                 logger,
		kubeClient:             o.kubeClient,
		namespace:              o.Pod.Namespace,
		ctx:                    ctx,
		cancelContext:          cancelContext,
		kubeInformerFactory:    kubeInformerFactory,
		anvilSideCarController: anvilSideCarController,
	}
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
