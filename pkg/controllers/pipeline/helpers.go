package pipeline

import (
	"github.com/golang/glog"
	"github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/controllers"
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	kubesmithInformers "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/kubesmith/v1"
	"github.com/sirupsen/logrus"
	appInformersv1 "k8s.io/client-go/informers/apps/v1"
	batchInformersv1 "k8s.io/client-go/informers/batch/v1"
	coreInformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewPipelineController(
	namespace string,
	maxRunningPipelines int,
	logger *logrus.Logger,
	kubeClient kubernetes.Interface,
	kubesmithClient kubesmithv1.KubesmithV1Interface,
	pipelineInformer kubesmithInformers.PipelineInformer,
	deploymentsInformer appInformersv1.DeploymentInformer,
	jobsInformer batchInformersv1.JobInformer,
	configMapsInformer coreInformersv1.ConfigMapInformer,
) controllers.Interface {
	c := &PipelineController{
		GenericController:   generic.NewGenericController("pipeline"),
		namespace:           namespace,
		maxRunningPipelines: maxRunningPipelines,
		logger:              logger,
		kubeClient:          kubeClient,
		kubesmithClient:     kubesmithClient,
		pipelineLister:      pipelineInformer.Lister(),
		deploymentLister:    deploymentsInformer.Lister(),
		jobLister:           jobsInformer.Lister(),
		configMapLister:     configMapsInformer.Lister(),
	}

	c.SyncHandler = c.processPipeline
	c.CacheSyncWaiters = append(
		c.CacheSyncWaiters,
		pipelineInformer.Informer().HasSynced,
		deploymentsInformer.Informer().HasSynced,
		jobsInformer.Informer().HasSynced,
		configMapsInformer.Informer().HasSynced,
	)

	// setup event handlers for pipelines
	pipelineInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pipeline := obj.(*v1.Pipeline)

				key, err := cache.MetaNamespaceKeyFunc(pipeline)
				if err != nil {
					glog.Errorf("Error creating queue key, item not added to queue; name: %s", pipeline.Name)
					return
				}

				c.Queue.Add(key)
			},
			UpdateFunc: func(oldObj, updatedObj interface{}) {
				oldPipeline := oldObj.(*v1.Pipeline)
				updatedPipeline := updatedObj.(*v1.Pipeline)

				key, err := cache.MetaNamespaceKeyFunc(updatedPipeline)
				if err != nil {
					glog.Errorf("Error updating queue key, item not added to queue; name: %s", updatedPipeline.Name)
					return
				}

				// create a tmp logger
				tmpLogger := c.logger.WithFields(logrus.Fields{
					"Name":       updatedPipeline.Name,
					"Namespace":  updatedPipeline.Namespace,
					"Phase":      updatedPipeline.Status.Phase,
					"StageIndex": updatedPipeline.Status.StageIndex,
				})

				// if the phase changed, react
				if updatedPipeline.Status.Phase != oldPipeline.Status.Phase {
					tmpLogger.Info("queueing pipeline: phase changed")
					c.Queue.Add(key)
					return
				}

				// if the phase is "running" and the stageIndex changed, react
				isRunningPhase := (updatedPipeline.Status.Phase == v1.PipelinePhaseRunning)
				stageIndexChanged := (updatedPipeline.Status.StageIndex != oldPipeline.Status.StageIndex)
				if isRunningPhase && stageIndexChanged {
					tmpLogger.Info("queueing pipeline: stage index advanced")
					c.Queue.Add(key)
				}
			},
		},
	)

	return c
}
