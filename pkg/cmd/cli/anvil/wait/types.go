package wait

import kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"

type Options struct {
	S3            OptionsS3
	FlagFile      OptionsFlagFile
	ArtifactPaths string
	Archive       OptionsArchive

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
}

type OptionsArchive struct {
	FileName string
	FilePath string
}
