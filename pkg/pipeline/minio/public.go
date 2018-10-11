package minio

import (
	"fmt"

	"github.com/kubesmith/kubesmith/pkg/pipeline/minio/templates"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *MinioServer) GetResourceName() string {
	return fmt.Sprintf("%s-minio-server", m.ResourcePrefix)
}

func (m *MinioServer) CreateSecret() error {
	resourceName := m.GetResourceName()
	secret, err := m.kubeClient.CoreV1().Secrets(m.Namespace).Get(resourceName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioSecret(resourceName, m.ResourceLabels)
			secret, err = m.kubeClient.CoreV1().Secrets(m.Namespace).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create minio secret")
			}

			m.minioSecret = secret
			return nil
		}

		return errors.Wrap(err, "could not get existing minio secret")
	}

	m.minioSecret = secret
	return nil
}

func (m *MinioServer) CreateDeployment() error {
	resourceName := m.GetResourceName()
	deployment, err := m.kubeClient.AppsV1().Deployments(m.Namespace).Get(resourceName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioDeployment(resourceName, m.ResourceLabels, *m.minioSecret)
			deployment, err = m.kubeClient.AppsV1().Deployments(m.Namespace).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create minio deployment")
			}

			m.minioDeployment = deployment
			return nil
		}

		return errors.Wrap(err, "could not get existing minio deployment")
	}

	m.minioDeployment = deployment
	return nil
}

func (m *MinioServer) CreateService() error {
	resourceName := m.GetResourceName()
	service, err := m.kubeClient.CoreV1().Services(m.Namespace).Get(resourceName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioService(resourceName, m.ResourceLabels)
			service, err = m.kubeClient.CoreV1().Services(m.Namespace).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create minio service")
			}

			m.minioService = service
			return nil
		}

		return errors.Wrap(err, "could not get existing minio service")
	}

	m.minioService = service
	return nil
}

func (m *MinioServer) Create() error {
	if err := m.CreateSecret(); err != nil {
		return err
	}

	if err := m.CreateDeployment(); err != nil {
		return err
	}

	return m.CreateService()
}

func (m *MinioServer) DeleteSecret() error {
	return m.kubeClient.CoreV1().Secrets(m.Namespace).Delete(m.GetResourceName(), &metav1.DeleteOptions{})
}

func (m *MinioServer) DeleteDeployment() error {
	return m.kubeClient.AppsV1().Deployments(m.Namespace).Delete(m.GetResourceName(), &metav1.DeleteOptions{})
}

func (m *MinioServer) DeleteService() error {
	return m.kubeClient.CoreV1().Services(m.Namespace).Delete(m.GetResourceName(), &metav1.DeleteOptions{})
}

func (m *MinioServer) Delete() error {
	if err := m.DeleteService(); err != nil {
		return err
	}

	if err := m.DeleteDeployment(); err != nil {
		return err
	}

	return m.DeleteSecret()
}
