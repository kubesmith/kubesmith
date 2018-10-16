package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	jobTemplates "github.com/kubesmith/kubesmith/pkg/pipeline/jobs/templates"
	"github.com/kubesmith/kubesmith/pkg/pipeline/minio"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (p *PipelineExecutor) canRunAnotherPipeline() (bool, error) {
	pl := p.GetCachedPipeline()
	pipelines, err := p.pipelineLister.Pipelines(pl.GetNamespace()).List(labels.Everything())
	if err != nil {
		return false, errors.Wrap(err, "could not list pipelines")
	}

	currentlyRunning := 0
	for _, pipeline := range pipelines {
		if pipeline.IsRunning() {
			currentlyRunning++
		}
	}

	if currentlyRunning < p.MaxRunningPipelines {
		return true, nil
	}

	return false, nil
}

func (p *PipelineExecutor) patchPipeline(pipeline api.Pipeline) (*api.Pipeline, error) {
	cached := p.GetCachedPipeline()
	patchType, patchBytes, err := pipeline.GetPatchFromOriginal(cached)
	if err != nil {
		return nil, err
	}

	updated, err := p.kubesmithClient.Pipelines(cached.GetNamespace()).Patch(cached.GetName(), patchType, patchBytes)
	if err != nil {
		return nil, err
	}

	p._cachedPipeline = *updated
	return updated.DeepCopy(), nil
}

func (p *PipelineExecutor) processEmptyPhasePipeline() error {
	p.logger.Info("validating pipeline...")
	if err := p.Validate(); err != nil {
		p.logger.Error(errors.Wrap(err, "could not validate pipeline"))

		p.logger.Info("marking pipeline as failed")
		pl := p.GetCachedPipeline()
		pl.SetPipelineToFailed(err.Error())
		if _, err := p.patchPipeline(pl); err != nil {
			p.logger.Error(errors.Wrap(err, "could not mark pipeline as failed"))
		}

		return err
	}

	p.logger.Info("finished validating pipeline")
	p.logger.Info("marking pipeline as queued...")
	pl := p.GetCachedPipeline()
	pl.SetPipelineToQueued()
	if _, err := p.patchPipeline(pl); err != nil {
		p.logger.Error(errors.Wrap(err, "could not set pipeline to running"))
		return err
	}
	p.logger.Info("marked pipeline as queued")

	return nil
}

func (p *PipelineExecutor) processQueuedPipeline() error {
	p.logger.Info("checking to see if we can run another pipeline in this namespace")
	canRunAnotherPipeline, err := p.canRunAnotherPipeline()
	if err != nil {
		p.logger.Error(errors.Wrap(err, "could not check to see if we could run another pipeline"))
		return err
	}

	if !canRunAnotherPipeline {
		p.logger.Warn("cannot run another pipeline in this namespace")
		return nil
	}

	p.logger.Info("marking pipeline as running...")
	pl := p.GetCachedPipeline()
	pl.SetPipelineToRunning()
	if _, err := p.patchPipeline(pl); err != nil {
		p.logger.Error(errors.Wrap(err, "could not set pipeline to running"))
		return err
	}

	p.logger.Info("marked pipeline as running")
	return nil
}

func (p *PipelineExecutor) processRunningPipeline() error {
	pl := p.GetCachedPipeline()
	p.logger.Info("ensuring minio server resources are scheduled...")
	minioServer := minio.NewMinioServer(
		pl.GetNamespace(),
		pl.GetResourcePrefix(),
		pl.GetResourceLabels(),
		p.logger,
		p.kubeClient,
		p.deploymentLister,
	)

	if err := minioServer.Create(); err != nil {
		p.logger.Error(errors.Wrap(err, "could not ensure minio server resources are scheduled"))
		return err
	}

	p.logger.Info("minio server resources are scheduled")
	p.logger.Info("waiting for minio server to be available...")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	minioServerAvailable := make(chan bool, 1)
	go minioServer.WaitForAvailability(ctx, 5, minioServerAvailable)

	if available := <-minioServerAvailable; !available {
		return errors.New("minio server is not available")
	}

	p.logger.Info("minio server is available!")
	p.logger.Info("ensuring jobs are scheduled")
	for index, job := range pl.GetExpandedJobsForCurrentStage() {
		jobIndex := index + 1
		logger := p.logger.WithField("JobIndex", jobIndex)

		logger.Info("scheduling job...")
		if err := p.scheduleJob(job, jobIndex, minioServer, logger); err != nil {
			err = errors.Wrap(err, "could not schedule job")
			logger.Error(err)
			return err
		}

		logger.Info("scheduled job!")
	}

	p.logger.Info("jobs are scheduled")
	return nil
}

func (p *PipelineExecutor) getJobResourceName(jobIndex int) string {
	pl := p.GetCachedPipeline()

	return fmt.Sprintf(
		"%s-stage-%d-job-%d",
		pl.GetResourcePrefix(),
		p._cachedPipeline.Status.StageIndex,
		jobIndex,
	)
}

func (p *PipelineExecutor) ensureJobConfigMapExists(name string, configMapData map[string]string, logger logrus.FieldLogger) error {
	pl := p.GetCachedPipeline()

	logger.Info("checking to see if configmap for job exists...")
	if _, err := p.configMapLister.ConfigMaps(pl.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			resource := jobTemplates.GetJobConfigMap(name, pl.GetResourceLabels(), configMapData)
			_, err := p.kubeClient.CoreV1().ConfigMaps(pl.GetNamespace()).Create(&resource)

			if err != nil {
				return errors.Wrap(err, "could not create configmap for job")
			}

			logger.Info("configmap for job created")
			return nil
		}

		return errors.Wrap(err, "could not find configmap for job")
	}

	logger.Info("configmap for job exists")
	return nil
}

func (p *PipelineExecutor) getPipelineJobConfigMapData(job api.PipelineSpecJob) map[string]string {
	if len(job.ConfigMapData) > 0 {
		return job.ConfigMapData
	}

	return map[string]string{"pipeline-script.sh": strings.Join(job.Runner, "\n")}
}

func (p *PipelineExecutor) getPipelineJobCommand(job api.PipelineSpecJob) []string {
	if len(job.Command) > 0 {
		return job.Command
	}

	return []string{"/bin/sh", "-x", "/kubesmith/scripts/pipeline-script.sh"}
}

func (p *PipelineExecutor) getPipelineJobArgs(job api.PipelineSpecJob) []string {
	if len(job.Args) > 0 {
		return job.Args
	}

	return []string{}
}

func (p *PipelineExecutor) scheduleJob(
	job api.PipelineSpecJob,
	jobIndex int,
	minioServer *minio.MinioServer,
	logger logrus.FieldLogger,
) error {
	// build a name for these resources
	pl := p.GetCachedPipeline()
	name := p.getJobResourceName(jobIndex)
	configMapData := p.getPipelineJobConfigMapData(job)

	// ensure the configMap exists (if it's needed)
	if len(configMapData) > 0 {
		if err := p.ensureJobConfigMapExists(name, configMapData, logger); err != nil {
			return err
		}
	}

	// now, ensure the job exists
	logger.Info("checking to see if job exists...")
	if _, err := p.jobLister.Jobs(pl.GetNamespace()).Get(name); err != nil {
		if apierrors.IsNotFound(err) {
			resource := jobTemplates.GetJob(
				name,
				job.Image,
				p.getPipelineJobCommand(job),
				p.getPipelineJobArgs(job),
				pl.GetResourceLabels(),
			)

			if _, err := p.kubeClient.BatchV1().Jobs(pl.GetNamespace()).Create(&resource); err != nil {
				return errors.Wrap(err, "could not create job")
			}

			logger.Info("job created")
			return nil
		}

		return errors.Wrap(err, "could not find job")
	}

	logger.Info("job exists")
	return nil
}

// cleanup all resources associated with the pipeline
func (p *PipelineExecutor) processFinishedPipeline() error {
	p.logger.Info("todo: processing finished pipeline")
	return nil
}
