package wait

import kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"

type Options struct {
	S3            OptionsS3
	FlagFile      OptionsFlagFile
	ArtifactPaths string
	ArchiveFile   OptionsArchiveFile

	client kubesmithClient.Interface
}

type OptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
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
