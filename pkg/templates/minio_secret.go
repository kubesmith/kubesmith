package templates

import (
	"github.com/kubesmith/kubesmith/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetMinioSecret(
	name, accessKey, secretKey string,
	labels map[string]string,
) corev1.Secret {
	return corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		StringData: map[string]string{
			accessKey: utils.GenerateRandomString(16),
			secretKey: utils.GenerateRandomString(32),
		},
	}
}
