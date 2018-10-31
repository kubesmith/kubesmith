package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineServiceAccount(pipeline api.Pipeline) corev1.ServiceAccount {
	return corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipeline.GetResourcePrefix(),
			Namespace: pipeline.GetNamespace(),
			Labels:    pipeline.GetLabels(),
		},
	}
}
