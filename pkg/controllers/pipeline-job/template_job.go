package pipelinejob

import (
	"strconv"

	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertEnvironentToEnvVar(environment map[string]string) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	for key, value := range environment {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return env
}

func GetJob(
	name, image, s3Host string,
	s3Port int,
	s3BucketName, s3SecretName, repoPath string,
	annotations, environment map[string]string,
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
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:            "setup-workspace",
							Image:           "kubesmith/kubesmith:0.1",
							ImagePullPolicy: "Always",
							Command:         []string{"kubesmith", "anvil", "extract"},
							Args:            []string{"--logtostderr", "-v", "2"},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "S3_HOST",
									Value: s3Host,
								},
								corev1.EnvVar{
									Name:  "S3_PORT",
									Value: strconv.Itoa(s3Port),
								},
								corev1.EnvVar{
									Name:  "S3_BUCKET_NAME",
									Value: "workspace",
								},
								corev1.EnvVar{
									Name: "S3_ACCESS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: s3SecretName,
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
												Name: s3SecretName,
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
									Name:  "LOCAL_PATH",
									Value: "/kubesmith/workspace",
								},
								corev1.EnvVar{
									Name:  "REMOTE_ARCHIVE_PATHS",
									Value: "repo.tar.gz",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "workspace",
									MountPath: "/kubesmith/workspace",
								},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:       "pipeline-job",
							Image:      image,
							Command:    command,
							Args:       args,
							WorkingDir: repoPath,
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "scripts",
									MountPath: "/kubesmith/scripts",
									ReadOnly:  false,
								},
								corev1.VolumeMount{
									Name:      "workspace",
									MountPath: repoPath,
								},
								corev1.VolumeMount{
									Name:      "artifacts",
									MountPath: "/kubesmith/artifacts",
								},
							},
							Env: convertEnvironentToEnvVar(environment),
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
