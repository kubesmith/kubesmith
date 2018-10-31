package templates

import (
	"fmt"
	"strconv"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetJobCloneRepoArchiveWorkspaceContainer(pipeline api.Pipeline) corev1.Container {
	s3UseSSL := "false"
	if pipeline.Spec.Workspace.Storage.S3.UseSSL == true {
		s3UseSSL = "true"
	}

	return corev1.Container{
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
				Value: pipeline.Spec.Workspace.Storage.S3.Host,
			},
			corev1.EnvVar{
				Name:  "S3_PORT",
				Value: strconv.Itoa(pipeline.Spec.Workspace.Storage.S3.Port),
			},
			corev1.EnvVar{
				Name:  "S3_BUCKET_NAME",
				Value: pipeline.Spec.Workspace.Storage.S3.BucketName,
			},
			corev1.EnvVar{
				Name: "S3_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: pipeline.Spec.Workspace.Storage.S3.Credentials.Secret.Name,
						},
						Key: pipeline.Spec.Workspace.Storage.S3.Credentials.Secret.AccessKeyKey,
					},
				},
			},
			corev1.EnvVar{
				Name: "S3_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: pipeline.Spec.Workspace.Storage.S3.Credentials.Secret.Name,
						},
						Key: pipeline.Spec.Workspace.Storage.S3.Credentials.Secret.SecretKeyKey,
					},
				},
			},
			corev1.EnvVar{
				Name:  "S3_USE_SSL",
				Value: s3UseSSL,
			},
			corev1.EnvVar{
				Name:  "S3_PATH",
				Value: fmt.Sprintf("%s/repo", pipeline.GetResourcePrefix()),
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
	}
}
