package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline"
	informers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/cache"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
	//
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}
	o.client = client

	kubeClient, err := f.KubeClient()
	if err != nil {
		return err
	}
	o.kubeClient = kubeClient

	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	var wg sync.WaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		glog.Infof("Received signal %s, shutting down", sig)
		cancelFunc()
	}()

	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(o.client, 0, informers.WithNamespace("kubesmith"))
	pipelineController := pipeline.NewPipelineController(
		o.client.KubesmithV1(),
		sharedInformerFactory.Kubesmith().V1().Pipelines(),
	)

	wg.Add(1)
	go func() {
		pipelineController.Run(ctx, 1)
		wg.Done()
	}()

	// SHARED INFORMERS HAVE TO BE STARTED AFTER ALL CONTROLLERS
	go sharedInformerFactory.Start(ctx.Done())

	cache.WaitForCacheSync(ctx.Done(), sharedInformerFactory.Kubesmith().V1().Pipelines().Informer().HasSynced)

	<-ctx.Done()

	glog.Info("Waiting for all controllers to shut down gracefully")
	wg.Wait()

	return nil
}
