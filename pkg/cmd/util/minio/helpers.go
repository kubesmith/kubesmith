package minio

import (
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/golang/glog"
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func getMinioSecrets(labels map[string]string) corev1.Secret {
	return corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "minio-server-credentials-" + generateRandomString(8),
			Labels: labels,
		},
		StringData: map[string]string{
			"access-key": generateRandomString(16),
			"secret-key": generateRandomString(32),
		},
	}
}

func getMinioDeployment(labels map[string]string, secret corev1.Secret) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "minio-server-" + generateRandomString(8),
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:        int32Ptr(1),
			MinReadySeconds: 5,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
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
											Key: "access-key",
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
											Key: "secret-key",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 9000,
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

func getMinioService(labels map[string]string) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "minio-" + generateRandomString(8),
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 9000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 9000,
					},
				},
			},
			Selector: labels,
		},
	}
}

func removeMinioSecretsForPipeline(pipeline *api.Pipeline, client kubernetes.Interface) error {
	secrets, err := client.Core().Secrets(pipeline.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("PipelineName=%s", pipeline.Name),
	})

	if err != nil {
		glog.V(1).Info("could not fetch secrets for the pipeline")
		return err
	}

	for _, secret := range secrets.Items {
		if err := client.Core().Secrets(pipeline.Namespace).Delete(secret.Name, &metav1.DeleteOptions{}); err != nil {
			glog.V(1).Info("could not delete secret for the pipeline")
			return err
		}
	}

	return nil
}

func removeMinioDeploymentsForPipeline(pipeline *api.Pipeline, client kubernetes.Interface) error {
	deployments, err := client.Apps().Deployments(pipeline.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("PipelineName=%s", pipeline.Name),
	})

	if err != nil {
		glog.V(1).Info("could not fetch deployments for the pipeline")
		return err
	}

	for _, deployment := range deployments.Items {
		if err := client.Apps().Deployments(pipeline.Namespace).Delete(deployment.Name, &metav1.DeleteOptions{}); err != nil {
			glog.V(1).Info("could not delete deployment for the pipeline")
			return err
		}
	}

	return nil
}

func removeMinioServicesForPipeline(pipeline *api.Pipeline, client kubernetes.Interface) error {
	services, err := client.Core().Services(pipeline.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("PipelineName=%s", pipeline.Name),
	})

	if err != nil {
		glog.V(1).Info("could not fetch services for the pipeline")
		return err
	}

	for _, service := range services.Items {
		if err := client.Core().Services(pipeline.Namespace).Delete(service.Name, &metav1.DeleteOptions{}); err != nil {
			glog.V(1).Info("could not delete service for the pipeline")
			return err
		}
	}

	return nil
}

func DeleteMinioServerForPipeline(pipeline *api.Pipeline, client kubernetes.Interface) error {
	if err := removeMinioSecretsForPipeline(pipeline, client); err != nil {
		return err
	}

	if err := removeMinioDeploymentsForPipeline(pipeline, client); err != nil {
		return err
	}

	if err := removeMinioServicesForPipeline(pipeline, client); err != nil {
		return err
	}

	return nil
}

func CreateMinioServerForPipeline(pipeline *api.Pipeline, client kubernetes.Interface) error {
	labels := map[string]string{
		"PipelineName": pipeline.Name,
	}

	secret := getMinioSecrets(labels)
	if _, err := client.Core().Secrets(pipeline.Namespace).Create(&secret); err != nil {
		glog.V(1).Info("could not create minio secret")
		return err
	}

	deployment := getMinioDeployment(labels, secret)
	if _, err := client.Apps().Deployments(pipeline.Namespace).Create(&deployment); err != nil {
		glog.V(1).Info("could not create minio deployment")
		return err
	}

	service := getMinioService(labels)
	if _, err := client.Core().Services(pipeline.Namespace).Create(&service); err != nil {
		glog.V(1).Info("could not create minio service")
		return err
	}

	return nil
}

func generateRandomString(s int, letters ...string) string {
	randomFactor := make([]byte, 1)
	_, err := rand.Read(randomFactor)
	if err != nil {
		return ""
	}

	mathrand.Seed(time.Now().UnixNano() * int64(randomFactor[0]))
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	if len(letters) > 0 {
		letterRunes = []rune(letters[0])
	}

	b := make([]rune, s)
	for i := range b {
		b[i] = letterRunes[mathrand.Intn(len(letterRunes))]
	}

	return string(b)
}

func int32Ptr(i int32) *int32 {
	return &i
}
