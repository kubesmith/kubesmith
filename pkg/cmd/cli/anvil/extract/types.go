package extract

import (
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/sirupsen/logrus"
)

type Options struct {
	S3        OptionsS3
	LocalPath string

	logger logrus.FieldLogger
}

type OptionsS3 struct {
	Host       string
	Port       int
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
	Path       string

	client *s3.S3Client
}
