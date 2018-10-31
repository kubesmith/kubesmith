package templates

import (
	"github.com/kubesmith/kubesmith/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetMinioDeployment(
	name, serviceAccount, accessKey, secretKey string,
	port int32,
	labels map[string]string,
	secret corev1.Secret,
) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:        utils.Int32Ptr(1),
			MinReadySeconds: 5,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccount,
					Containers: []corev1.Container{
						{
							Name:            "minio-server",
							Image:           "minio/minio",
							ImagePullPolicy: "IfNotPresent",
							Command: []string{
								"minio",
								"server",
								"/data",
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "MINIO_ACCESS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secret.Name,
											},
											Key: accessKey,
										},
									},
								},
								corev1.EnvVar{
									Name: "MINIO_SECRET_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secret.Name,
											},
											Key: secretKey,
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "storage",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "storage",
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
