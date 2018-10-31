package sidecar

import (
	"context"

	"github.com/kubesmith/kubesmith/pkg/controllers"
	kubesmithClient "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/kubesmith/kubesmith/pkg/s3"
	"github.com/sirupsen/logrus"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Sidecar        OptionsSidecar
	Pod            OptionsPod
	S3             OptionsS3
	ArchiveFile    OptionsArchiveFile
	TimeoutSeconds int

	client     kubesmithClient.Interface
	kubeClient kubernetes.Interface
}

type OptionsPod struct {
	Name      string
	Namespace string
}

type OptionsSidecar struct {
	Name string
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

type OptionsArchiveFile struct {
	Name string
	Path string
}

type Server struct {
	options    *Options
	logger     *logrus.Logger
	kubeClient kubernetes.Interface
	namespace  string

	ctx                 context.Context
	cancelContext       context.CancelFunc
	kubeInformerFactory kubeInformers.SharedInformerFactory

	anvilSideCarController controllers.Interface
}
