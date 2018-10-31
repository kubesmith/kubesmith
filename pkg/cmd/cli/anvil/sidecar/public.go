package sidecar

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Pod.Name, "pod-name", "", "The name of the pod to monitor")
	env.BindEnvToFlag("pod-name", flags)
	flags.StringVar(&o.Pod.Namespace, "pod-namespace", "", "The namespace of the pod to monitor")
	env.BindEnvToFlag("pod-namespace", flags)
	flags.StringVar(&o.Sidecar.Name, "sidecar-name", "", "The name of the sidecar container")
	env.BindEnvToFlag("sidecar-name", flags)
	flags.StringVar(&o.S3.Host, "s3-host", "minio.default.svc", "The host where the s3 server is running")
	env.BindEnvToFlag("s3-host", flags)
	flags.IntVar(&o.S3.Port, "s3-port", 9000, "The s3 port that artifacts will be synced to/from")
	env.BindEnvToFlag("s3-port", flags)
	flags.StringVar(&o.S3.AccessKey, "s3-access-key", "", "The s3 access key that is used for authentication when making requests against the s3 server (required)")
	env.BindEnvToFlag("s3-access-key", flags)
	flags.StringVar(&o.S3.SecretKey, "s3-secret-key", "", "The s3 secret key that is used for authentication when making requests against the s3 server (required)")
	env.BindEnvToFlag("s3-secret-key", flags)
	flags.StringVar(&o.S3.BucketName, "s3-bucket-name", "artifacts", "The s3 bucket where the archive of artifacts will be uploaded to")
	env.BindEnvToFlag("s3-bucket-name", flags)
	flags.StringVar(&o.S3.Path, "s3-path", "", "The s3 path (inside the specified bucket) where the archive of artifacts will be uploaded to")
	env.BindEnvToFlag("s3-path", flags)
	flags.BoolVar(&o.S3.UseSSL, "s3-use-ssl", true, "Indicates whether to use SSL when connecting to the s3 server")
	env.BindEnvToFlag("s3-use-ssl", flags)
	flags.StringVar(&o.ArchiveFile.Name, "archive-file-name", "artifacts.tar.gz", "The name of the compressed file that will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-name", flags)
	flags.StringVar(&o.ArchiveFile.Path, "archive-file-path", os.TempDir(), "The directory where the compressed file will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-path", flags)
	flags.IntVar(&o.TimeoutSeconds, "timeout-seconds", 300, "The length of time before the sidecar exits due to timeout")
	env.BindEnvToFlag("timeout-seconds", flags)
	flags.IntVar(&o.WatchIntervalSeconds, "watch-interval-seconds", 1, "The interval (in seconds) at which the sidecar will check for updates on the specified pod")
	env.BindEnvToFlag("watch-interval-seconds", flags)
	flags.StringVar(&o.SuccessArtifactPaths, "success-artifact-paths", "", "A comma-separated list of artifact paths that anvil will look to upload when the pod succeeds; please note that golang glob patterns are supported")
	env.BindEnvToFlag("success-artifact-paths", flags)
	flags.StringVar(&o.FailArtifactPaths, "fail-artifact-paths", "", "A comma-separated list of artifact paths that anvil will look to upload when the pod fails; please note that golang glob patterns are supported")
	env.BindEnvToFlag("fail-artifact-paths", flags)
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
	// make sure pod name is specified
	if o.Pod.Name == "" {
		return fmt.Errorf("invalid pod name")
	}

	// make sure pod namespace is specified
	if o.Pod.Namespace == "" {
		return fmt.Errorf("invalid pod namespace")
	}

	// make sure sidecar name is specified
	if o.Sidecar.Name == "" {
		return fmt.Errorf("invalid sidecar name")
	}

	// make sure a valid archive extension was specified
	if !archive.IsValidArchiveExtension(o.getArchiveFilePath()) {
		return archive.GetInvalidFileFormatError()
	}

	// access key length: https://github.com/minio/minio/blob/master/docs/config/README.md
	if len(o.S3.AccessKey) < 3 {
		return fmt.Errorf("Invalid s3 access key")
	}

	// secret key length: https://github.com/minio/minio/blob/master/docs/config/README.md
	if len(o.S3.SecretKey) < 8 {
		return fmt.Errorf("Invalid s3 secret key")
	}

	// make sure the archive file path exists
	if _, err := os.Stat(o.ArchiveFile.Path); os.IsNotExist(err) {
		return fmt.Errorf("Archive file path does not exist")
	}

	// ensure the pod exists
	if _, err := o.podLister.Pods(o.Pod.Namespace).Get(o.Pod.Name); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("pod was not found")
		}

		return errors.Wrap(err, "could not retrieve pod from lister")
	}

	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	// setup our kube client
	kubeClient, err := f.KubeClient()
	if err != nil {
		return err
	}
	o.kubeClient = kubeClient

	// make sure our pod namespace is filled in
	if o.Pod.Namespace == "" {
		ns, err := getNamespaceFromServiceAccount()
		if err == nil {
			o.Pod.Namespace = ns
		} else {
			o.Pod.Namespace = "default"
		}
	}

	// make sure our sidecar has a name
	if o.Sidecar.Name == "" {
		o.Sidecar.Name = "kubesmith"
	}

	// create an s3 client
	s3Client, err := s3.NewS3Client(o.S3.Host, o.S3.Port, o.S3.AccessKey, o.S3.SecretKey, o.S3.UseSSL)
	if err != nil {
		return err
	} else {
		o.S3.client = s3Client
	}

	// create a logger
	o.logger = logrus.New().WithField("name", "sidecar")

	// create a context
	o.ctx, o.cancelContext = context.WithTimeout(context.Background(), time.Second*time.Duration(o.TimeoutSeconds))

	// call the cancelContext function if we receive a interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// cancel the context if a signal is fired
	go func() {
		sig := <-sigs
		o.logger.Infof("Received signal %s, shutting down", sig)
		o.cancelContext()
	}()

	// setup a new informer
	o.kubeInformerFactory = kubeInformers.NewSharedInformerFactoryWithOptions(
		o.kubeClient,
		0,
		kubeInformers.WithNamespace(o.Pod.Namespace),
		kubeInformers.WithTweakListOptions(func(listOptions *metav1.ListOptions) {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", o.Pod.Name)
		}),
	)

	// pass our context to the kube informer
	go o.kubeInformerFactory.Start(o.ctx.Done())

	// setup the cache sync waiter
	cache.WaitForCacheSync(
		o.ctx.Done(),
		o.kubeInformerFactory.Core().V1().Pods().Informer().HasSynced,
	)

	// create our pod lister
	o.podLister = o.kubeInformerFactory.Core().V1().Pods().Lister()

	// finally, return
	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	// start some routines to watch things
	go o.waitForContextToFinish()
	go o.checkPodListerCacheForUpdates()

	// now, wait for the context to timeout, complete or be cancelled...
	<-o.ctx.Done()

	// finally, return no errors
	return nil
}
