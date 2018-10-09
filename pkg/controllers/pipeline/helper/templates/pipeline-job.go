package templates

import (
	"fmt"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildJobFromPipelineJob(
	pipeline *api.Pipeline,
	jobIndex int,
	job api.PipelineSpecJob,
	configMap corev1.ConfigMap,
	minioService corev1.Service,
	minioSecret corev1.Secret,
) batchv1.Job {
	artifactPaths := []string{}
	flagFileName := ""

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipeline-job-" + generateRandomString(8),
			Labels: map[string]string{
				"PipelineName": pipeline.Name,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						corev1.Container{
							Name:  job.Name,
							Image: job.Image,
							Command: []string{
								"bash", "/kubesmith/scripts/job.sh",
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "scripts",
									MountPath: "/kubesmith/scripts",
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
									Value: fmt.Sprintf("%s.%s.svc", minioService.Name, minioService.Namespace),
								},
								corev1.EnvVar{
									Name:  "S3_PORT",
									Value: "9000",
								},
								corev1.EnvVar{
									Name:  "S3_BUCKET_NAME",
									Value: pipeline.Name,
								},
								corev1.EnvVar{
									Name: "S3_ACCESS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: minioSecret.Name,
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
												Name: minioSecret.Name,
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
									Value: fmt.Sprintf("artifacts-stage-%d-job-%d.tar.gz", pipeline.Status.StageIndex, jobIndex),
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
									Value: "/kubesmith/workspace/" + flagFileName,
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
							Name: "scripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap.Name,
									},
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
