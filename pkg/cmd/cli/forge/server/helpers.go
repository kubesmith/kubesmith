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
	"github.com/kubesmith/kubesmith/pkg/controllers/forge"
	"github.com/kubesmith/kubesmith/pkg/controllers/job"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline"
	pipelinejob "github.com/kubesmith/kubesmith/pkg/controllers/pipeline-job"
	pipelinestage "github.com/kubesmith/kubesmith/pkg/controllers/pipeline-stage"
	kubesmithInformers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kubeInformers "k8s.io/client-go/informers"
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

	// setup a logger for this pipeline controller
	logger := logrus.New()

	// setup our informers
	kubesmithInformerFactory := kubesmithInformers.NewSharedInformerFactoryWithOptions(
		o.client,
		0,
		kubesmithInformers.WithNamespace(o.Namespace),
	)

	kubeInformerFactory := kubeInformers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		0,
		kubeInformers.WithNamespace(o.Namespace),
	)

	// setup our controllers
	forgeController := forge.NewForgeController(
		logger,
		o.kubeClient,
		o.client.KubesmithV1(),
		kubesmithInformerFactory.Kubesmith().V1().Forges(),
	)

	pipelineController := pipeline.NewPipelineController(
		o.MaxRunningPipelines,
		logger,
		o.kubeClient,
		o.client.KubesmithV1(),
		kubesmithInformerFactory.Kubesmith().V1().Pipelines(),
		kubesmithInformerFactory.Kubesmith().V1().PipelineStages(),
		kubeInformerFactory.Core().V1().Secrets(),
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Core().V1().Services(),
		kubeInformerFactory.Batch().V1().Jobs(),
	)

	pipelineStageController := pipelinestage.NewPipelineStageController(
		logger,
		o.kubeClient,
		o.client.KubesmithV1(),
		kubesmithInformerFactory.Kubesmith().V1().PipelineStages(),
	)

	pipelineJobController := pipelinejob.NewPipelineJobController(
		logger,
		o.kubeClient,
		o.client.KubesmithV1(),
		kubesmithInformerFactory.Kubesmith().V1().PipelineJobs(),
	)

	jobController := job.NewJobController(
		logger,
		o.kubeClient,
		o.client.KubesmithV1(),
		kubeInformerFactory.Batch().V1().Jobs(),
	)

	// finally, return the server
	return &Server{
		options:    o,
		client:     o.client,
		logger:     logger,
		kubeClient: o.kubeClient,
		namespace:  o.Namespace,

		ctx:                      ctx,
		cancelContext:            cancelContext,
		kubesmithInformerFactory: kubesmithInformerFactory,
		kubeInformerFactory:      kubeInformerFactory,

		forgeController:         forgeController,
		pipelineController:      pipelineController,
		pipelineStageController: pipelineStageController,
		pipelineJobController:   pipelineJobController,
		jobController:           jobController,
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
