package extract

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/s3"
	uuid "github.com/satori/go.uuid"
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
	flags.StringVar(&o.LocalPath, "local-path", os.TempDir(), "The local path to a folder where the remote archive will be extracted")
	env.BindEnvToFlag("local-path", flags)
	flags.StringVar(&o.RemoteArchivePaths, "remote-archive-paths", "artifacts.tar.gz", "The paths of the remote archive(s) that will be downloaded from the specified s3 bucket")
	env.BindEnvToFlag("remote-archive-paths", flags)
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
	// ensure the local path exists
	glog.V(1).Infoln("Ensuring local file path exists...")
	if err := os.MkdirAll(o.LocalPath, os.ModePerm); err != nil {
		return err
	}

	// make sure the remote archives exist before continuing
	glog.V(1).Infoln("Ensuring remote archives exist...")
	for _, file := range o.GetRemoteArchivePaths() {
		exists, err := o.S3.client.FileExists(o.S3.BucketName, file)

		if err != nil {
			return err
		} else if !exists {
			glog.Exitf("Remote file s3://%s/%s does not exist", o.S3.BucketName, file)
		} else {
			glog.V(1).Infof("Found s3://%s/%s", o.S3.BucketName, file)
		}
	}

	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	s3Client, err := s3.NewS3Client(o.S3.Host, o.S3.Port, o.S3.AccessKey, o.S3.SecretKey, o.S3.UseSSL)
	if err != nil {
		return err
	}

	o.S3.client = s3Client
	return nil
}

func (o *Options) GetRemoteArchivePaths() []string {
	paths := []string{}
	files := strings.Split(o.RemoteArchivePaths, ",")

	for _, file := range files {
		file = strings.TrimSpace(file)
		paths = append(paths, file)
	}

	return paths
}

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	// download the archive files and extract them to the local file path
	for _, remoteFilePath := range o.GetRemoteArchivePaths() {
		localFilePath := fmt.Sprintf("%s%s%s", os.TempDir(), uuid.NewV4(), path.Ext(remoteFilePath))

		// download the archive first
		glog.V(1).Infof("Downloading file from s3://%s/%s to %s ... ", o.S3.BucketName, remoteFilePath, localFilePath)
		if err := o.S3.client.DownloadFile(o.S3.BucketName, remoteFilePath, localFilePath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// now, extract the archive to our localFilePath
		glog.V(1).Infoln("Downloaded; extracting archive...")
		if err := archive.ExtractArchive(localFilePath, o.LocalPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// finally, remove the downloaded archive
		glog.V(1).Infoln("Extracted; cleaning up downloaded archive...")
		if err := os.Remove(localFilePath); err != nil {
			fmt.Printf("Could not clean up downloaded archive at %s ...\n", localFilePath)
		}
	}

	// finally, we're done!
	glog.V(1).Infof("Successfully extracted remote file archives to %s!", o.LocalPath)
	return nil
}
