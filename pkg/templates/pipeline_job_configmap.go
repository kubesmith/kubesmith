package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineJobConfigMap(job api.PipelineJob) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   job.GetName(),
			Labels: job.GetLabels(),
		},
		Data: job.GetConfigMapData(),
	}
}
