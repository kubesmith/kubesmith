package sidecar

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	if !archive.IsValidArchiveExtension(o.GetArchiveFilePath()) {
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

	if o.Pod.Namespace == "" {
		ns, err := getNamespaceFromServiceAccount()
		if err == nil {
			o.Pod.Namespace = ns
		} else {
			o.Pod.Namespace = "default"
		}
	}

	if o.Sidecar.Name == "" {
		o.Sidecar.Name = "kubesmith"
	}

	s3Client, err := s3.NewS3Client(o.S3.Host, o.S3.Port, o.S3.AccessKey, o.S3.SecretKey, o.S3.UseSSL)
	if err != nil {
		return err
	}

	o.S3.client = s3Client
	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	server := NewServer(o)

	if err := server.run(); err != nil {
		return err
	}

	return nil
}

func (o *Options) GetArchiveFilePath() string {
	path := strings.TrimRight(o.ArchiveFile.Path, "/")
	path = strings.TrimRight(path, "\\")

	return fmt.Sprintf(
		"%s%s%s",
		path,
		string(os.PathSeparator),
		o.ArchiveFile.Name,
	)
}
