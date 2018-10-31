package templates

import (
	"fmt"
	"strings"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetJobCloneRepoCheckoutRepoContainer(pipeline api.Pipeline) corev1.Container {
	return corev1.Container{
		Name:    "checkout-git-repo",
		Image:   "alpine/git",
		Command: []string{"/bin/sh", "-xc"},
		Args: []string{
			strings.Join([]string{
				fmt.Sprintf("git clone %s /git/workspace", pipeline.Spec.Workspace.Repo.URL),
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
	}
}
