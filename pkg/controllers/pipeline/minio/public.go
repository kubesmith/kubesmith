package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/kubesmith/kubesmith/pkg/templates"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *MinioServer) GetResourceName() string {
	return fmt.Sprintf("%s-minio-server", m.resourcePrefix)
}

func (m *MinioServer) CreateSecret() error {
	resourceName := m.GetResourceName()
	secret, err := m.secretLister.Secrets(m.namespace).Get(resourceName)

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioSecret(
				resourceName,
				MINIO_DEFAULT_ACCESS_KEY_KEY,
				MINIO_DEFAULT_SECRET_KEY_KEY,
				m.resourceLabels,
			)

			secret, err = m.kubeClient.CoreV1().Secrets(m.namespace).Create(&resource)

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
	deployment, err := m.deploymentLister.Deployments(m.namespace).Get(resourceName)

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioDeployment(
				resourceName,
				m.resourcePrefix,
				MINIO_DEFAULT_ACCESS_KEY_KEY,
				MINIO_DEFAULT_SECRET_KEY_KEY,
				MINIO_DEFAULT_PORT,
				m.resourceLabels,
				*m.minioSecret,
			)

			deployment, err = m.kubeClient.AppsV1().Deployments(m.namespace).Create(&resource)

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

func (m *MinioServer) GetBucketName() string {
	return fmt.Sprintf("%s-%s", m.resourcePrefix, MINIO_DEFAULT_BUCKET_NAME)
}

func (m *MinioServer) CreateService() error {
	resourceName := m.GetResourceName()
	service, err := m.serviceLister.Services(m.namespace).Get(resourceName)

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := templates.GetMinioService(
				resourceName,
				MINIO_DEFAULT_PORT,
				m.resourceLabels,
			)

			service, err = m.kubeClient.CoreV1().Services(m.namespace).Create(&resource)

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

func (m *MinioServer) getDeleteOptions() *metav1.DeleteOptions {
	propagationPolicy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
}

func (m *MinioServer) DeleteSecret() error {
	if err := m.kubeClient.CoreV1().Secrets(m.namespace).Delete(m.GetResourceName(), m.getDeleteOptions()); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *MinioServer) DeleteDeployment() error {
	if err := m.kubeClient.AppsV1().Deployments(m.namespace).Delete(m.GetResourceName(), m.getDeleteOptions()); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *MinioServer) DeleteService() error {
	if err := m.kubeClient.CoreV1().Services(m.namespace).Delete(m.GetResourceName(), m.getDeleteOptions()); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
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

func (m *MinioServer) WaitForAvailability(
	ctx context.Context,
	secondsInterval int,
	minioServerAvailable chan bool,
) {
	// rework this function once we have a better understanding of how to wait
	// for a specific thing
	namespace := m.minioDeployment.GetNamespace()
	name := m.minioDeployment.GetName()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				minioServerAvailable <- false
				return
			}
		default:
			deployment, err := m.deploymentLister.Deployments(namespace).Get(name)
			if err != nil {
				m.logger.Info(errors.Wrap(err, "could not fetch minio deployment"))
			}

			if deployment.Status.ReadyReplicas == 1 {
				minioServerAvailable <- true
				return
			}
		}

		// make sure to sleep so we don't hammer the api server
		time.Sleep(time.Second * time.Duration(secondsInterval))
	}
}

func (m *MinioServer) GetServiceHost() (string, error) {
	if m.minioService == nil {
		return "", errors.New("minio service has not been created")
	}

	host := fmt.Sprintf("%s.%s.svc", m.minioService.GetName(), m.minioService.GetNamespace())
	return host, nil
}
