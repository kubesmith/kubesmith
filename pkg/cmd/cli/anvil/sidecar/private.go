package sidecar

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
	"github.com/kubesmith/kubesmith/pkg/cmd/util/artifacts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (o *Options) checkPodListerCacheForUpdates() {
	// wait out the interval
	time.Sleep(time.Second * time.Duration(o.WatchIntervalSeconds))

	// retrieve the pod from cache
	pod, err := o.podLister.Pods(o.Pod.Namespace).Get(o.Pod.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			o.logger.Info("deleted pod detected; exiting gracefully")
			os.Exit(0)
		}

		o.logger.Fatal(errors.Wrap(err, "could not retrieve pod from lister"))
	}

	// check container's pod statuses and then keep calling itself
	o.checkPodContainerStatuses(*pod)
	o.checkPodListerCacheForUpdates()
}

func (o *Options) checkPodContainerStatuses(pod corev1.Pod) {
	o.logger.Info("scanning container statuses")

	allContainersHaveExited := true
	exitCode := 0

	for _, status := range pod.Status.ContainerStatuses {
		if o.Sidecar.Name == status.Name {
			continue
		}

		if status.State.Terminated == nil {
			allContainersHaveExited = false
			continue
		} else if status.State.Terminated.ExitCode > 0 {
			exitCode = 1
		}
	}

	if allContainersHaveExited {
		if exitCode == 0 {
			if err := o.processSuccessfulPod(); err != nil {
				o.logger.Info(err)
			}
		} else {
			if err := o.processFailedPod(); err != nil {
				o.logger.Info(err)
			}
		}

		os.Exit(exitCode)
	}

	o.logger.Info("skipping; containers still running")
}

func (o *Options) waitForContextToFinish() {
	for {
		select {
		case <-o.ctx.Done():
			if o.ctx.Err() == context.DeadlineExceeded {
				o.logger.Info("time limit exceeded; exiting sidecar")
				os.Exit(1)
			}

			return
		case <-time.After(time.Second):
			break
		}
	}
}

func (o *Options) processSuccessfulPod() error {
	o.logger.Info("processing successful pod")

	// detect any artifacts that were expected to be created
	detectedArtifacts := artifacts.DetectFromCSV(o.SuccessArtifactPaths)
	if len(detectedArtifacts) == 0 {
		return errors.New("No artifacts were detected")
	}

	// compress the artifacts and then upload them
	return o.compressArtifactsAndUpload(detectedArtifacts)
}

func (o *Options) processFailedPod() error {
	o.logger.Info("processing failed pod")

	// detect any artifacts that were expected to be created
	detectedArtifacts := artifacts.DetectFromCSV(o.FailArtifactPaths)
	if len(detectedArtifacts) == 0 {
		return errors.New("No artifacts were detected")
	}

	// compress the artifacts and then upload them
	return o.compressArtifactsAndUpload(detectedArtifacts)
}

func (o *Options) compressArtifactsAndUpload(artifacts []string) error {
	filePath := o.getArchiveFilePath()
	o.logger.Infof("Detected artifact(s); Compressing to %s ...", filePath)

	if err := archive.CreateArchive(filePath, artifacts); err != nil {
		return errors.Wrapf(err, "could not create artifact archive at %s", filePath)
	}

	// upload the compressed archive to s3
	o.logger.Info("Compressed archive; Uploading to S3...")
	if err := o.S3.client.UploadFileToBucket(filePath, o.S3.BucketName, o.S3.Path); err != nil {
		return errors.Wrap(err, "Could not upload artifacts to S3")
	}

	// cleanup!
	o.logger.Info("Deleting local archive to cleanup...")
	if err := os.Remove(filePath); err != nil {
		o.logger.Infof("Could not clean up local archive at %s ...", filePath)
	}

	// finally... we're done
	o.logger.Info("Successfully compressed and uploaded artifacts!")
	return nil
}

func (o *Options) getArchiveFilePath() string {
	path := strings.TrimRight(o.ArchiveFile.Path, "/")
	path = strings.TrimRight(path, "\\")

	return fmt.Sprintf(
		"%s%s%s",
		path,
		string(os.PathSeparator),
		o.ArchiveFile.Name,
	)
}
