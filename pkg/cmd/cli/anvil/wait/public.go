package wait

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/client"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/artifacts"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/env"
	watcher "github.com/kubesmith/kubesmith/pkg/cmd/util/file-watcher"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
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
	flags.StringVar(&o.S3.BucketName, "s3-bucket-name", "artifacts", "The s3 bucket where the archive of artifacts will be uploaded to")
	env.BindEnvToFlag("s3-bucket-name", flags)
	flags.BoolVar(&o.S3.UseSSL, "s3-use-ssl", true, "Indicates whether to use SSL when connecting to the s3 server")
	env.BindEnvToFlag("s3-use-ssl", flags)
	flags.StringVar(&o.FlagFile.Path, "flag-file-path", "", "The file anvil will watch for until it exists")
	env.BindEnvToFlag("flag-file-path", flags)
	flags.IntVar(&o.FlagFile.WatchInterval, "flag-file-watch-interval", 1, "The interval (in seconds) in which anvil will check for the flag file")
	env.BindEnvToFlag("flag-file-watch-interval", flags)
	flags.IntVar(&o.FlagFile.WatchTimeout, "flag-file-watch-timeout", 300, "The number of seconds until waiting for the flag file to exist times out")
	env.BindEnvToFlag("flag-file-watch-timeout", flags)
	flags.StringVar(&o.ArtifactPaths, "artifact-paths", "", "A comma-separated list of paths that anvil will look for when compressing the archive to sync; please note that golang glob patterns are supported")
	env.BindEnvToFlag("artifact-paths", flags)
	flags.StringVar(&o.ArchiveFile.Name, "archive-file-name", "artifacts.tar.gz", "The name of the compressed file that will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-name", flags)
	flags.StringVar(&o.ArchiveFile.Path, "archive-file-path", os.TempDir(), "The directory where the compressed file will be created if artifacts are found")
	env.BindEnvToFlag("archive-file-path", flags)
}

func (o *Options) Validate(c *cobra.Command, args []string, f client.Factory) error {
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

	// make sure the flag file path is specified
	if o.FlagFile.Path == "" {
		return fmt.Errorf("Invalid flag file path")
	}

	return nil
}

func (o *Options) Complete(args []string, f client.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	o.client = client
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

func (o *Options) Run(c *cobra.Command, f client.Factory) error {
	// create a golang channel for when the file exists
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(o.FlagFile.WatchTimeout))
	flagFileCreated := make(chan bool, 1)

	// start watching for the file to exist
	glog.V(1).Infof("Watching for flag file to exist at: %s for %d second(s) ...", o.FlagFile.Path, o.FlagFile.WatchTimeout)
	go watcher.WatchForFile(ctx, o.FlagFile.Path, o.FlagFile.WatchInterval, flagFileCreated)

	// hang on the golang channel until the watcher detects the flag file
	if created := <-flagFileCreated; !created {
		glog.Exitln("Flag file was not created in time")
	}
	glog.V(1).Infoln("Flag file detected! Detecting artifacts...")

	// detect any artifacts that were expected to be created
	detectedArtifacts := artifacts.DetectFromCSV(o.ArtifactPaths)
	if len(detectedArtifacts) == 0 {
		glog.Exitln("No artifacts were detected before the timeout ...")
	}

	// compress all of the files into a tarball
	filePath := o.GetArchiveFilePath()
	glog.V(1).Infof("Detected artifact(s); Compressing to %s ...", filePath)
	if err := archive.CreateArchive(filePath, detectedArtifacts); err != nil {
		glog.V(1).Infof("Could not create artifact archive at %s ...", filePath)
		glog.Exit(err)
	}

	// create a new s3 client
	s3Client, err := s3.NewS3Client(o.S3.Host, o.S3.Port, o.S3.AccessKey, o.S3.SecretKey, o.S3.UseSSL)
	if err != nil {
		return err
	}

	// upload the compressed tarball to s3
	glog.V(1).Infoln("Compressed tarball; Uploading to S3...")
	if err := s3Client.UploadFileToBucket(filePath, o.S3.BucketName); err != nil {
		glog.V(1).Infoln("Could not upload artifacts to S3...")
		glog.Exit(err)
	}

	// cleanup!
	glog.V(1).Infoln("Deleting local archive to cleanup...")
	if err := os.Remove(filePath); err != nil {
		glog.Infof("Could not clean up local archive at %s ...\n", filePath)
	}

	// finally... we're done!
	glog.V(1).Infoln("Successfully uploaded!")
	return nil
}
