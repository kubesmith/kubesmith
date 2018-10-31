package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineStage(name string, pipeline api.Pipeline) api.PipelineStage {
	return api.PipelineStage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "PipelineStage",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: pipeline.GetLabels(),
		},
		Spec: api.PipelineStageSpec{
			Workspace: api.PipelineStageWorkspace{
				Path:    pipeline.GetWorkspacePath(),
				Storage: pipeline.Spec.Workspace.Storage,
			},
			Jobs: pipeline.GetExpandedJobsForCurrentStage(),
		},
	}
}
