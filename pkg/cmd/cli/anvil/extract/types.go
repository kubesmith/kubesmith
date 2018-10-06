package extract

import kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"

type Options struct {
	S3                 OptionsS3
	LocalPath          string
	RemoteArchivePaths string

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
