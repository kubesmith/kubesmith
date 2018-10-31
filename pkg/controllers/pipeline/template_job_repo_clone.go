package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRepoCloneJob(
	name, pipelineName string,
	repo api.WorkspaceRepo,
	s3Config api.WorkspaceStorageS3,
	labels map[string]string,
) batchv1.Job {
	s3UseSSL := "false"
	if s3Config.UseSSL == true {
		s3UseSSL = "true"
	}

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: utils.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:    "setup-ssh",
							Image:   "alpine/git",
							Command: []string{"/bin/sh", "-xc"},
							Args: []string{
								strings.Join([]string{
									"cp /ssh/id_rsa /root/.ssh/id_rsa",
									"chmod 0400 /root/.ssh/id_rsa",
									"echo \"Host github.com\" > /root/.ssh/config",
									"echo \"    StrictHostKeyChecking no\" >> /root/.ssh/config",
									"echo \"    LogLevel error\" >> /root/.ssh/config",
								}, "; "),
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "root",
									MountPath: "/root/.ssh",
								},
								corev1.VolumeMount{
									Name:      "ssh",
									MountPath: "/ssh",
								},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "checkout-git-repo",
							Image:   "alpine/git",
							Command: []string{"/bin/sh", "-xc"},
							Args: []string{
								strings.Join([]string{
									fmt.Sprintf("git clone %s /git/workspace", repo.URL),
									"rm -rf /git/workspace/.git",
									"ls -la /git/workspace",
								}, "; "),
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "root",
									MountPath: "/root/.ssh",
								},
								corev1.VolumeMount{
									Name:      "workspace",
									MountPath: "/git",
								},
							},
						},
						corev1.Container{
							Name:            "copy-workspace-to-s3",
							Image:           "kubesmith/kubesmith",
							ImagePullPolicy: "Always",
							Command:         []string{"kubesmith", "anvil", "sidecar"},
							Args:            []string{"--logtostderr", "-v", "2"},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								corev1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								corev1.EnvVar{
									Name:  "SIDECAR_NAME",
									Value: "copy-workspace-to-s3",
								},
								corev1.EnvVar{
									Name:  "S3_HOST",
									Value: s3Config.Host,
								},
								corev1.EnvVar{
									Name:  "S3_PORT",
									Value: strconv.Itoa(s3Config.Port),
								},
								corev1.EnvVar{
									Name:  "S3_BUCKET_NAME",
									Value: s3Config.BucketName,
								},
								corev1.EnvVar{
									Name: "S3_ACCESS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: s3Config.Credentials.Secret.Name,
											},
											Key: s3Config.Credentials.Secret.AccessKeyKey,
										},
									},
								},
								corev1.EnvVar{
									Name: "S3_SECRET_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: s3Config.Credentials.Secret.Name,
											},
											Key: s3Config.Credentials.Secret.SecretKeyKey,
										},
									},
								},
								corev1.EnvVar{
									Name:  "S3_USE_SSL",
									Value: s3UseSSL,
								},
								corev1.EnvVar{
									Name:  "S3_PATH",
									Value: fmt.Sprintf("%s/repo", pipelineName),
								},
								corev1.EnvVar{
									Name:  "ARCHIVE_FILE_NAME",
									Value: "repo.tar.gz",
								},
								corev1.EnvVar{
									Name:  "ARCHIVE_FILE_PATH",
									Value: "/artifacts",
								},
								corev1.EnvVar{
									Name:  "SUCCESS_ARTIFACT_PATHS",
									Value: "/git/workspace/**",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "workspace",
									MountPath: "/git",
								},
								corev1.VolumeMount{
									Name:      "artifacts",
									MountPath: "/artifacts",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "root",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
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
						corev1.Volume{
							Name: "ssh",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: repo.SSH.Secret.Name,
									Items: []corev1.KeyToPath{
										corev1.KeyToPath{
											Key:  repo.SSH.Secret.Key,
											Path: "id_rsa",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
