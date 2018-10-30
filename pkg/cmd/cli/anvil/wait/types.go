package wait

import "github.com/kubesmith/kubesmith/pkg/s3"

type Options struct {
	S3            OptionsS3
	FlagFile      OptionsFlagFile
	ArtifactPaths string
	ArchiveFile   OptionsArchiveFile
}

type OptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	Path       string
	UseSSL     bool

	client *s3.S3Client
}

type OptionsFlagFile struct {
	Path          string
	WatchInterval int
	WatchTimeout  int
}

type OptionsArchiveFile struct {
	Name string
	Path string
}
