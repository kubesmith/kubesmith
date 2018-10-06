package client

import (
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	clientset "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

type Factory interface {
	// BindFlags binds common flags (--kubeconfig, --namespace) to the passed-in FlagSet.
	BindFlags(flags *pflag.FlagSet)

	// Client returns a KubesmithClient. It uses the following priority to specify the cluster
	// configuration: --kubeconfig flag, KUBECONFIG environment variable, in-cluster configuration.
	Client() (clientset.Interface, error)

	// KubeClient returns a Kubernetes client. It uses the following priority to specify the cluster
	// configuration: --kubeconfig flag, KUBECONFIG environment variable, in-cluster configuration.
	KubeClient() (kubernetes.Interface, error)

	Namespace() string
}

type factory struct {
	flags       *pflag.FlagSet
	kubeconfig  string
	kubecontext string
	baseName    string
	namespace   string
}

func NewFactory(baseName string) Factory {
	f := &factory{
		flags:    pflag.NewFlagSet("", pflag.ContinueOnError),
		baseName: baseName,
	}

	if f.namespace == "" {
		f.namespace = v1.DefaultNamespace
	}

	f.flags.StringVar(&f.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file to use to talk to the Kubernetes apiserver. If unset, try the environment variable KUBECONFIG, as well as in-cluster configuration")
	f.flags.StringVarP(&f.namespace, "namespace", "n", f.namespace, "The namespace in which Kubesmith should operate")
	f.flags.StringVar(&f.kubecontext, "kubecontext", "", "The context to use to talk to the Kubernetes apiserver. If unset defaults to whatever your current-context is (kubectl config current-context)")

	return f
}

func (f *factory) BindFlags(flags *pflag.FlagSet) {
	flags.AddFlagSet(f.flags)
}

func (f *factory) Client() (clientset.Interface, error) {
	clientConfig, err := Config(f.kubeconfig, f.kubecontext, f.baseName)
	if err != nil {
		return nil, err
	}

	kubesmithClient, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return kubesmithClient, nil
}

func (f *factory) KubeClient() (kubernetes.Interface, error) {
	clientConfig, err := Config(f.kubeconfig, f.kubecontext, f.baseName)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return kubeClient, nil
}

func (f *factory) Namespace() string {
	return f.namespace
}
