package extract

import (
	"github.com/kubesmith/kubesmith/pkg/cmd/util/s3"
)

type Options struct {
	S3                 OptionsS3
	LocalPath          string
	RemoteArchivePaths string
}

type OptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool

	client *s3.S3Client
}
