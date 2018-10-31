package sidecar

import (
	"context"
	"os"
	"time"

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
			o.processDeletedPod()
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
		} else {
			exitCode = 1
		}
	}

	if allContainersHaveExited {
		if exitCode == 0 {
			o.processSuccessfulPod()
			os.Exit(0)
		}

		o.processFailedPod()
		os.Exit(exitCode)
	}

	o.logger.Info("skipping; containers still running")
}

func (o *Options) waitForContextToFinish() {
	for {
		select {
		case <-o.ctx.Done():
			if o.ctx.Err() == context.DeadlineExceeded {
				o.contextTimedOut()
			}

			return
		case <-time.After(time.Second):
			break
		}
	}
}

func (o *Options) processSuccessfulPod() {
	o.logger.Info("processing successful pod")
}

func (o *Options) processFailedPod() {
	o.logger.Info("processing failed pod")
}

func (o *Options) processDeletedPod() {
	o.logger.Info("processing deleted pod")
}

func (o *Options) contextTimedOut() {
	o.logger.Info("timed out")
}
