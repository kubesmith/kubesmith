package templates

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPipelineJob(job api.PipelineJob) batchv1.Job {
	template := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   job.GetName(),
			Labels: job.GetLabels(),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: utils.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					InitContainers: []corev1.Container{
						GetPipelineJobSetupWorkspaceInitContainer(job),
					},
					Containers: []corev1.Container{
						GetPipelineJobPrimaryContainer(job),
						GetPipelineJobAnvilSidecarContainer(job),
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "scripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: job.GetName(),
									},
									DefaultMode: utils.Int32Ptr(0777),
								},
							},
						},
						corev1.Volume{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						corev1.Volume{
							Name: "artifacts",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	previousStageName := job.GetPreviousPipelineStageName()
	if previousStageName != "" {
		template.Spec.Template.Spec.InitContainers = append(template.Spec.Template.Spec.InitContainers,
			GetPipelineJobDownloadArtifactsInitContainer(job))
	}

	return template
}
