package anvil

import (
	"os"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewWaitCommand(f client.Factory, use string) *cobra.Command {
	o := NewWaitOptions()

	c := &cobra.Command{
		Use:   use,
		Short: "Waits for a flag file to exist and optionally uploads any detected artifacts to a remote s3 bucket",
		Long:  "Waits for a flag file to exist and optionally uploads any detected artifacts to a remote s3 bucket",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Complete(args, f))
			cmd.CheckError(o.Validate(c, args, f))
			cmd.CheckError(o.Run(c, f))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

type WaitOptions struct {
	S3            WaitOptionsS3
	FlagFile      WaitOptionsFlagFile
	ArtifactPaths string
	Archive       WaitOptionsArchive

	client kubesmithClient.Interface
}

type WaitOptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

type WaitOptionsFlagFile struct {
	Path          string
	WatchInterval int
}

type WaitOptionsArchive struct {
	FileName string
	FilePath string
}

func NewWaitOptions() *WaitOptions {
	return &WaitOptions{
		S3: WaitOptionsS3{},
	}
}

func (o *WaitOptions) BindFlags(flags *pflag.FlagSet) {
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
	flags.BoolVar(&o.S3.UseSSL, "s3-use-ssl", true, "Indicates whether to use SSL when connecting to the s3 server")
	env.BindEnvToFlag("s3-use-ssl", flags)
	flags.StringVar(&o.FlagFile.Path, "flag-file-path", "", "The file anvil will watch for until it exists")
	env.BindEnvToFlag("flag-file-path", flags)
	flags.IntVar(&o.FlagFile.WatchInterval, "flag-file-watch-interval", 1, "The interval (in seconds) in which anvil will check for the flag file")
	env.BindEnvToFlag("flag-file-watch-interval", flags)
	flags.StringVar(&o.ArtifactPaths, "artifact-paths", "", "A comma-separated list of paths that anvil will look for when compressing the archive to sync; please note that golang glob patterns are supported")
	env.BindEnvToFlag("artifact-paths", flags)
	flags.StringVar(&o.Archive.FileName, "archive-file-name", "artifacts.tar.gz", "The name of the compressed file that will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-name", flags)
	flags.StringVar(&o.Archive.FilePath, "archive-file-path", os.TempDir(), "The directory where the compressed file will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-path", flags)
}

func (o *WaitOptions) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *WaitOptions) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	o.client = client
	return nil
}

func (o *WaitOptions) Run(c *cobra.Command, f client.Factory) error {
	return nil
}
