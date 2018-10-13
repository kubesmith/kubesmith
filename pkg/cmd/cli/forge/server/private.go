package server

import (
	"sync"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (s *Server) runControllers() error {
	var wg sync.WaitGroup

	// start the pipeline controller
	wg.Add(1)
	go func() {
		s.pipelineController.Run(s.ctx, 1)
		wg.Done()
	}()

	// start the shared informers after all of our controllers
	go s.kubesmithInformerFactory.Start(s.ctx.Done())
	go s.kubeInformerFactory.Start(s.ctx.Done())

	// setup the cache sync waiter
	cache.WaitForCacheSync(
		s.ctx.Done(),
		s.kubesmithInformerFactory.Kubesmith().V1().Pipelines().Informer().HasSynced,
		s.kubeInformerFactory.Apps().V1().Deployments().Informer().HasSynced,
		s.kubeInformerFactory.Batch().V1().Jobs().Informer().HasSynced,
		s.kubeInformerFactory.Core().V1().ConfigMaps().Informer().HasSynced,
	)

	<-s.ctx.Done()

	glog.Info("Waiting for all controllers to shut down gracefully")
	wg.Wait()

	return nil
}

func (s *Server) run() error {
	if err := s.namespaceExists(s.namespace); err != nil {
		return err
	}

	if err := s.runControllers(); err != nil {
		return err
	}

	return nil
}

func (s *Server) namespaceExists(namespace string) error {
	glog.V(1).Infof("Checking existence of %s", namespace)

	if _, err := s.kubeClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		return errors.WithStack(err)
	}

	glog.V(1).Info("Namespace exists")
	return nil
}
