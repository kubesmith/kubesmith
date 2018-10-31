package templates

import (
	"fmt"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJobCloneRepo(
	pipeline api.Pipeline,
) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-clone-repo", pipeline.GetResourcePrefix()),
			Labels: pipeline.GetLabels(),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: utils.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: pipeline.GetResourcePrefix(),
					RestartPolicy:      "Never",
					InitContainers: []corev1.Container{
						GetJobCloneRepoSetupSSHInitContainer(),
					},
					Containers: []corev1.Container{
						GetJobCloneRepoCheckoutRepoContainer(pipeline),
						GetJobCloneRepoArchiveWorkspaceContainer(pipeline),
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
									SecretName: pipeline.Spec.Workspace.Repo.SSH.Secret.Name,
									Items: []corev1.KeyToPath{
										corev1.KeyToPath{
											Key:  pipeline.Spec.Workspace.Repo.SSH.Secret.Key,
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
