package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineJob(
	name string,
	stage api.PipelineStage,
	job api.PipelineJobSpecJob,
) api.PipelineJob {
	return api.PipelineJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "PipelineJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: stage.GetLabels(),
		},
		Spec: api.PipelineJobSpec{
			Workspace: api.PipelineJobWorkspace{
				Path:    stage.Spec.Workspace.Path,
				Storage: stage.Spec.Workspace.Storage,
			},
			Job: job,
		},
	}
}
