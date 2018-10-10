package helper

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/controllers/pipeline/helper/templates"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (p *PipelineHelper) createMinioServerSecret() error {
	glog.V(1).Info("checking to see if the minio server secret already exists for this pipeline")
	secret, err := p.kubeClient.CoreV1().Secrets(p.pipeline.Namespace).Get(p.getMinioServerResourceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(1).Info("minio server secret does not exist; creating...")
			resource := templates.GetMinioSecret(p.getMinioServerResourceName(), p.resourceLabels)
			secret, err = p.kubeClient.CoreV1().Secrets(p.pipeline.Namespace).Create(&resource)
			if err != nil {
				return errors.Wrap(err, "could not create minio server secret")
			}
		}

		return errors.Wrap(err, "could not get existing minio server secret")
	}

	p.minioSecret = secret

	return nil
}

func (p *PipelineHelper) createMinioServerDeployment() error {
	if err := p.createMinioServerSecret(); err != nil {
		return err
	}

	glog.V(1).Info("checking to see if the minio server deployment already exists for this pipeline")
	deployment, err := p.kubeClient.AppsV1().Deployments(p.pipeline.Namespace).Get(p.getMinioServerResourceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(1).Info("minio server deployment does not exist; creating...")
			resource := templates.GetMinioDeployment(p.getMinioServerResourceName(), p.resourceLabels, *p.minioSecret)
			deployment, err = p.kubeClient.AppsV1().Deployments(p.pipeline.Namespace).Create(&resource)
			if err != nil {
				return errors.Wrap(err, "could not create minio server deployment")
			}
		}

		return errors.Wrap(err, "could not get existing minio server deployment")
	}

	p.minioDeployment = deployment

	return nil
}

func (p *PipelineHelper) createMinioServerService() error {
	glog.V(1).Info("checking to see if the minio server service already exists for this pipeline")
	service, err := p.kubeClient.CoreV1().Services(p.pipeline.Namespace).Get(p.getMinioServerResourceName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(1).Info("creating minio server service")
			resource := templates.GetMinioService(p.getMinioServerResourceName(), p.resourceLabels)
			service, err = p.kubeClient.CoreV1().Services(p.pipeline.Namespace).Create(&resource)
			if err != nil {
				return errors.Wrap(err, "could not create minio server service")
			}
		}

		return errors.Wrap(err, "could not get existing minio server service")
	}

	p.minioService = service

	return nil
}

func (p *PipelineHelper) createMinioServer() error {
	if err := p.createMinioServerDeployment(); err != nil {
		return err
	}

	if err := p.createMinioServerService(); err != nil {
		return err
	}

	return nil
}

func (p *PipelineHelper) getMinioServerResourceName() string {
	return fmt.Sprintf("%s-minio-server", p.resourcePrefix)
}

func (p *PipelineHelper) deleteMinioServerSecret() error {
	return p.kubeClient.CoreV1().Secrets(p.pipeline.Namespace).Delete(p.getMinioServerResourceName(), &metav1.DeleteOptions{})
}

func (p *PipelineHelper) deleteMinioServerDeployment() error {
	return p.kubeClient.AppsV1().Deployments(p.pipeline.Namespace).Delete(p.getMinioServerResourceName(), &metav1.DeleteOptions{})
}

func (p *PipelineHelper) deleteMinioServerService() error {
	return p.kubeClient.CoreV1().Services(p.pipeline.Namespace).Delete(p.getMinioServerResourceName(), &metav1.DeleteOptions{})
}

func (p *PipelineHelper) deleteMinioServer() error {
	if err := p.deleteMinioServerService(); err != nil {
		return err
	}

	if err := p.deleteMinioServerDeployment(); err != nil {
		return err
	}

	if err := p.deleteMinioServerSecret(); err != nil {
		return err
	}

	return nil
}
