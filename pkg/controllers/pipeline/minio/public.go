package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/kubesmith/kubesmith/pkg/s3"
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
			resource := GetMinioSecret(resourceName, m.ResourceLabels)
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
	deployment, err := m.deploymentLister.Deployments(m.Namespace).Get(resourceName)

	if err != nil {
		if apierrors.IsNotFound(err) {
			resource := GetMinioDeployment(resourceName, m.ResourceLabels, *m.minioSecret)
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
			resource := GetMinioService(resourceName, m.ResourceLabels)
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

func (m *MinioServer) getDeleteOptions() *metav1.DeleteOptions {
	propagationPolicy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}
}

func (m *MinioServer) DeleteSecret() error {
	return m.kubeClient.CoreV1().Secrets(m.Namespace).Delete(m.GetResourceName(), m.getDeleteOptions())
}

func (m *MinioServer) DeleteDeployment() error {
	return m.kubeClient.AppsV1().Deployments(m.Namespace).Delete(m.GetResourceName(), m.getDeleteOptions())
}

func (m *MinioServer) DeleteService() error {
	return m.kubeClient.CoreV1().Services(m.Namespace).Delete(m.GetResourceName(), m.getDeleteOptions())
}

func (m *MinioServer) Delete() error {
	if err := m.DeleteService(); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	if err := m.DeleteDeployment(); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	if err := m.DeleteSecret(); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *MinioServer) WaitForAvailability(
	ctx context.Context,
	secondsInterval int,
	minioServerAvailable chan bool,
) {
	// rework this function once we have a better understanding of how to wait
	// for a specific thing
	namespace := m.minioDeployment.Namespace
	name := m.minioDeployment.Name

	for {
		select {
		case <-ctx.Done():
			minioServerAvailable <- false
			break
		default:
			deployment, err := m.deploymentLister.Deployments(namespace).Get(name)
			if err != nil {
				m.logger.Warn(errors.Wrap(err, "could not fetch minio deployment"))
			}

			if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
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

func (m *MinioServer) GetS3Client() (*s3.S3Client, error) {
	if m.minioService == nil {
		return nil, errors.New("minio service has not been created")
	} else if m.minioSecret == nil {
		return nil, errors.New("minio secret has not been created")
	}

	_, err := m.GetServiceHost()
	if err != nil {
		return nil, err
	}

	return s3.NewS3Client(
		// host,
		"127.0.0.1",
		int(m.minioService.Spec.Ports[0].Port),
		string(m.minioSecret.Data["access-key"]),
		string(m.minioSecret.Data["secret-key"]),
		false,
	)
}
