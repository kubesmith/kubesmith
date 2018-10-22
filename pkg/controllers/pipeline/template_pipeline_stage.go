package pipeline

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineStage(
	name string,
	storage api.WorkspaceStorage,
	labels map[string]string,
	jobs []api.PipelineJobSpecJob,
) api.PipelineStage {
	return api.PipelineStage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "PipelineStage",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: api.PipelineStageSpec{
			Workspace: api.PipelineStageWorkspace{
				Storage: storage,
			},
			Jobs: jobs,
		},
	}
}
