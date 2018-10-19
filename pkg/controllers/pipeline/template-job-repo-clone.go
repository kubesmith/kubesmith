package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRepoCloneJob(
	name, sshSecretName, sshSecretKey, repoURL, s3Host string,
	s3Port int,
	s3SecretName string,
	labels map[string]string,
) batchv1.Job {
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
									fmt.Sprintf("git clone %s /git/workspace", repoURL),
									"rm -rf /git/workspace/.git",
									"ls -la /git/workspace",
									"touch /git/kubesmith-ipc-flag",
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
							Image:           "kubesmith/kubesmith:0.1",
							ImagePullPolicy: "Always",
							Command:         []string{"kubesmith", "anvil", "wait"},
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
									Name:  "ARCHIVE_FILE_NAME",
									Value: "repo.tar.gz",
								},
								corev1.EnvVar{
									Name:  "ARCHIVE_FILE_PATH",
									Value: "/artifacts",
								},
								corev1.EnvVar{
									Name:  "ARTIFACT_PATHS",
									Value: "/git/workspace/**",
								},
								corev1.EnvVar{
									Name:  "FLAG_FILE_PATH",
									Value: "/git/kubesmith-ipc-flag",
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
									SecretName: sshSecretName,
									Items: []corev1.KeyToPath{
										corev1.KeyToPath{
											Key:  sshSecretKey,
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
