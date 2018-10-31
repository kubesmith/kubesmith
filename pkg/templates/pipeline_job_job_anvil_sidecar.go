package templates

import (
	"fmt"
	"strconv"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetPipelineJobJobAnvilSidecarContainer(job api.PipelineJob) corev1.Container {
	s3UseSSL := "false"
	if job.Spec.Workspace.Storage.S3.UseSSL == true {
		s3UseSSL = "true"
	}

	return corev1.Container{
		Name:            "anvil-sidecar",
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
				Value: "anvil-sidecar",
			},
			corev1.EnvVar{
				Name:  "S3_HOST",
				Value: job.Spec.Workspace.Storage.S3.Host,
			},
			corev1.EnvVar{
				Name:  "S3_PORT",
				Value: strconv.Itoa(job.Spec.Workspace.Storage.S3.Port),
			},
			corev1.EnvVar{
				Name:  "S3_BUCKET_NAME",
				Value: job.Spec.Workspace.Storage.S3.BucketName,
			},
			corev1.EnvVar{
				Name: "S3_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: job.Spec.Workspace.Storage.S3.Credentials.Secret.Name,
						},
						Key: job.Spec.Workspace.Storage.S3.Credentials.Secret.AccessKeyKey,
					},
				},
			},
			corev1.EnvVar{
				Name: "S3_SECRET_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: job.Spec.Workspace.Storage.S3.Credentials.Secret.Name,
						},
						Key: job.Spec.Workspace.Storage.S3.Credentials.Secret.SecretKeyKey,
					},
				},
			},
			corev1.EnvVar{
				Name:  "S3_USE_SSL",
				Value: s3UseSSL,
			},
			corev1.EnvVar{
				Name:  "S3_PATH",
				Value: fmt.Sprintf("%s/%s", job.GetPipelineName(), job.GetPipelineStageName()),
			},
			corev1.EnvVar{
				Name:  "ARCHIVE_FILE_NAME",
				Value: fmt.Sprintf("%s.tar.gz", job.GetPipelineJobName()),
			},
			corev1.EnvVar{
				Name:  "ARCHIVE_FILE_PATH",
				Value: "/kubesmith/artifacts",
			},
			corev1.EnvVar{
				Name:  "SUCCESS_ARTIFACT_PATHS",
				Value: strings.Join(job.GetSuccessArtifactPaths(), ","),
			},
			corev1.EnvVar{
				Name:  "FAIL_ARTIFACT_PATHS",
				Value: strings.Join(job.GetFailArtifactPaths(), ","),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "workspace",
				MountPath: job.Spec.Workspace.Path,
			},
			corev1.VolumeMount{
				Name:      "artifacts",
				MountPath: "/kubesmith/artifacts",
			},
		},
	}
}
