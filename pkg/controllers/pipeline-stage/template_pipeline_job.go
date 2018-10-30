package pipelinestage

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineJob(
	name, workspacePath string,
	labels map[string]string,
	storage api.WorkspaceStorage,
	job api.PipelineJobSpecJob,
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
		Spec: api.PipelineJobSpec{
			Workspace: api.PipelineJobWorkspace{
				Path:    workspacePath,
				Storage: storage,
			},
			Job: job,
		},
	}
}
