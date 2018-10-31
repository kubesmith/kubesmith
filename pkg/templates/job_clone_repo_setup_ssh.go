package templates

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func GetJobCloneRepoSetupSSHInitContainer() corev1.Container {
	return corev1.Container{
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
	}
}
