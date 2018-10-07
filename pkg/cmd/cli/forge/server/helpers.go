package server

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions"
	"github.com/spf13/cobra"
)

func NewCommand(f client.Factory) *cobra.Command {
	o := NewOptions()

	c := &cobra.Command{
		Use:   "server",
		Short: "Starts the forge server",
		Long:  "Starts the forge server",
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

	// setup some pipeline helpers
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(o.client, 0, informers.WithNamespace(o.Namespace))
	pipelineController := pipeline.NewPipelineController(
		o.Namespace,
		o.client.KubesmithV1(),
		sharedInformerFactory.Kubesmith().V1().Pipelines(),
	)

	// finally, return the server
	return &Server{
		options:    o,
		client:     o.client,
		kubeClient: o.kubeClient,
		namespace:  o.Namespace,

		ctx:                   ctx,
		cancelContext:         cancelContext,
		sharedInformerFactory: sharedInformerFactory,
		pipelineController:    pipelineController,
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
