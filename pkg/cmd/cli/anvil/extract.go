package anvil

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewExtractCommand(f client.Factory, use string) *cobra.Command {
	o := NewExtractOptions()

	c := &cobra.Command{
		Use:   use,
		Short: "Extracts remote s3 archive(s) locally",
		Long:  "Extracts remote s3 archive(s) locally",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Complete(args, f))
			cmd.CheckError(o.Validate(c, args, f))
			cmd.CheckError(o.Run(c, f))
		},
	}

	o.BindFlags(c.Flags())

	return c
}

type ExtractOptions struct {
	S3                 ExtractOptionsS3
	LocalPath          string
	RemoteArchivePaths string

	client kubesmithClient.Interface
}

type ExtractOptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

func NewExtractOptions() *ExtractOptions {
	return &ExtractOptions{
		S3: ExtractOptionsS3{},
	}
}

func (o *ExtractOptions) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.S3.Host, "s3-host", "minio.default.svc", "The host where the s3 server is running")
	env.BindEnvToFlag("s3-host", flags)
	flags.IntVar(&o.S3.Port, "s3-port", 9000, "The s3 port that artifacts will be synced to/from")
	env.BindEnvToFlag("s3-port", flags)
	flags.StringVar(&o.S3.AccessKey, "s3-access-key", "", "The s3 access key that is used for authentication when making requests against the s3 server (required)")
	env.BindEnvToFlag("s3-access-key", flags)
	flags.StringVar(&o.S3.SecretKey, "s3-secret-key", "", "The s3 secret key that is used for authentication when making requests against the s3 server (required)")
	env.BindEnvToFlag("s3-secret-key", flags)
	flags.StringVar(&o.S3.BucketName, "s3-bucket-name", "artifacts", "The s3 bucket where the archive of artifacts will be downloaded from")
	env.BindEnvToFlag("s3-bucket-name", flags)
	flags.BoolVar(&o.S3.UseSSL, "s3-use-ssl", true, "Indicates whether to use SSL when connecting to the s3 server")
	env.BindEnvToFlag("s3-use-ssl", flags)
	flags.StringVar(&o.LocalPath, "local-path", os.TempDir(), "The local path to a folder where the remote archive will be extracted")
	env.BindEnvToFlag("local-path", flags)
	flags.StringVar(&o.RemoteArchivePaths, "remote-archive-paths", "artifacts.tar.gz", "The paths of the remote archive(s) that will be downloaded from the specified s3 bucket")
	env.BindEnvToFlag("remote-archive-paths", flags)
}

func (o *ExtractOptions) Validate(c *cobra.Command, args []string, f client.Factory) error {
	return nil
}

func (o *ExtractOptions) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	o.client = client
	return nil
}

func (o *ExtractOptions) Run(c *cobra.Command, f client.Factory) error {
	spew.Dump(o.S3)

	return nil
}
