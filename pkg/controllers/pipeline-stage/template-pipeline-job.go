package pipelinestage

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineJob(
	name string,
	labels map[string]string,
	job api.PipelineJobSpec,
) api.PipelineJob {
	return api.PipelineJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "PipelineJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: job,
	}
}
