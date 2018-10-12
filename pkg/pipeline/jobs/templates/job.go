package templates

import (
	"fmt"
	"strings"

	"github.com/kubesmith/kubesmith/pkg/pipeline/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJob(
	labels map[string]string,
	pipelineName string,
	jobName, jobImage string,
	archiveFileName string,
	minioServiceHost, minioServerSecretName string,
	artifactPaths []string,
	commands []string,
) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: utils.Int32Ptr(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "pipeline-job",
							Image: jobImage,
							Command: []string{
								"/bin/sh", "-c",
							},
							Args: []string{
								strings.Join(commands, "; ") + ";",
							},
							VolumeMounts: []corev1.VolumeMount{
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
						corev1.Container{
							Name:            "kubesmith-anvil",
							Image:           "kubesmith/kubesmith:0.1",
							ImagePullPolicy: "Always",
							Command: []string{
								"kubesmith",
								"anvil",
								"wait",
								"--logtostderr",
								"-v",
								"2",
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "S3_HOST",
									Value: minioServiceHost,
								},
								corev1.EnvVar{
									Name:  "S3_PORT",
									Value: "9000",
								},
								corev1.EnvVar{
									Name:  "S3_BUCKET_NAME",
									Value: pipelineName,
								},
								corev1.EnvVar{
									Name: "S3_ACCESS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: minioServerSecretName,
											},
											Key: "access-key",
										},
									},
								},
								corev1.EnvVar{
									Name: "S3_SECRET_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: minioServerSecretName,
											},
											Key: "secret-key",
										},
									},
								},
								corev1.EnvVar{
									Name:  "S3_USE_SSL",
									Value: "false",
								},
								corev1.EnvVar{
									Name:  "ARCHIVE_FILE_NAME",
									Value: archiveFileName,
								},
								corev1.EnvVar{
									Name:  "ARCHIVE_FILE_PATH",
									Value: "/kubesmith/artifacts",
								},
								corev1.EnvVar{
									Name:  "ARTIFACT_PATHS",
									Value: strings.Join(artifactPaths, ","),
								},
								corev1.EnvVar{
									Name:  "FLAG_FILE_PATH",
									Value: fmt.Sprintf("/kubesmith/workspace/%s", pipelineName),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
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
