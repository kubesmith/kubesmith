package templates

import (
	"fmt"
	"strconv"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetPipelineJobJobDownloadArtifactsInitContainer(job api.PipelineJob) corev1.Container {
	s3UseSSL := "false"
	if job.Spec.Workspace.Storage.S3.UseSSL == true {
		s3UseSSL = "true"
	}

	return corev1.Container{
		Name:            "download-artifacts",
		Image:           "kubesmith/kubesmith",
		ImagePullPolicy: "Always",
		Command:         []string{"kubesmith", "anvil", "extract"},
		Args:            []string{"--logtostderr", "-v", "2"},
		Env: []corev1.EnvVar{
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
				Value: fmt.Sprintf("%s/%s", job.GetPipelineName(), job.GetPreviousPipelineStageName()),
			},
			corev1.EnvVar{
				Name:  "LOCAL_PATH",
				Value: "/kubesmith/workspace",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "workspace",
				MountPath: "/kubesmith/workspace",
			},
		},
	}
}
