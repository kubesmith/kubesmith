package extract

import (
	"fmt"
	"os"
	"path"

	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o *Options) BindFlags(flags *pflag.FlagSet) {
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
	flags.StringVar(&o.S3.Path, "s3-path", "artifacts.tar.gz", "The directory that contains archives which will be downloaded and extracted from the specified s3 bucket")
	env.BindEnvToFlag("s3-path", flags)
	flags.StringVar(&o.LocalPath, "local-path", os.TempDir(), "The local path to a folder where the remote archive will be extracted")
	env.BindEnvToFlag("local-path", flags)
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
	// ensure the local path exists
	if err := os.MkdirAll(o.LocalPath, os.ModePerm); err != nil {
		return err
	}

	// access key length: https://github.com/minio/minio/blob/master/docs/config/README.md
	if len(o.S3.AccessKey) < 3 {
		return fmt.Errorf("Invalid s3 access key")
	}

	// secret key length: https://github.com/minio/minio/blob/master/docs/config/README.md
	if len(o.S3.SecretKey) < 8 {
		return fmt.Errorf("Invalid s3 secret key")
	}

	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	s3Client, err := s3.NewS3Client(o.S3.Host, o.S3.Port, o.S3.AccessKey, o.S3.SecretKey, o.S3.UseSSL)
	if err != nil {
		return err
	}

	o.S3.client = s3Client
	o.logger = logrus.New().WithField("name", "extract")

	return nil
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	archives, err := o.getArchivesFromS3Path()
	if err != nil {
		return errors.Wrap(err, "could not retrieve archives from s3 path")
	} else if len(archives) == 0 {
		o.logger.Info("no artifacts detected in s3 path; exiting")
		os.Exit(0)
	}

	o.logger.Infof("detected %d archive(s)", len(archives))
	for _, remoteArchivePath := range archives {
		localFilePath := fmt.Sprintf("%s%s%s", os.TempDir(), uuid.NewV4(), path.Ext(remoteArchivePath))

		// download the archive first
		o.logger.Infof("downloading archive from s3://%s/%s to %s", o.S3.BucketName, remoteArchivePath, localFilePath)
		if err := o.S3.client.DownloadFile(o.S3.BucketName, remoteArchivePath, localFilePath); err != nil {
			return errors.Wrap(err, "could not download archive from s3")
		}

		// now, extract the archive to our local file path
		o.logger.Info("downloaded archive; extracting...")
		if err := archive.ExtractArchive(localFilePath, o.LocalPath); err != nil {
			return errors.Wrap(err, "could not extract archive")
		}

		// finally, remove the downloaded archive
		o.logger.Info("extracted archive; cleaning up...")
		if err := os.Remove(localFilePath); err != nil {
			o.logger.Infof("could not clean up downloaded archive at %s", localFilePath)
		}
	}

	// finally, we're done!
	o.logger.Infof("successfully downloaded and extracted remote file archives to %s", o.LocalPath)
	return nil
}
