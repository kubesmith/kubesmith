package pipeline

import (
	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJob(
	name, image string,
	annotations map[string]string,
	command, args []string,
	labels map[string]string,
) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: utils.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "pipeline-job",
							Image:   image,
							Command: command,
							Args:    args,
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "scripts",
									MountPath: "/kubesmith/scripts",
									ReadOnly:  false,
								},
								corev1.VolumeMount{
									Name:      "workspace",
									MountPath: "/kubesmith/workspace",
								},
								corev1.VolumeMount{
									Name:      "artifacts",
									MountPath: "/kubesmith/artifacts",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "scripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name,
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
}
