package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetPipelineJobJobPrimaryContainer(job api.PipelineJob) corev1.Container {
	return corev1.Container{
		Name:       "pipeline-job",
		Image:      job.Spec.Job.Image,
		Command:    job.Spec.Job.Command,
		Args:       job.Spec.Job.Args,
		WorkingDir: job.Spec.Workspace.Path,
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "scripts",
				MountPath: "/kubesmith/scripts",
				ReadOnly:  false,
			},
			corev1.VolumeMount{
				Name:      "workspace",
				MountPath: job.Spec.Workspace.Path,
			},
			corev1.VolumeMount{
				Name:      "artifacts",
				MountPath: "/kubesmith/artifacts",
			},
		},
		Env: convertEnvironentToEnvVar(job.Spec.Job.Environment),
	}
}
