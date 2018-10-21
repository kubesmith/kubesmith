package pipeline

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineStage(
	name string,
	labels map[string]string,
	jobs []api.PipelineJobSpec,
) api.PipelineStage {
	return api.PipelineStage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "pipelinestage",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: api.PipelineStageSpec{
			Jobs: jobs,
		},
	}
}
