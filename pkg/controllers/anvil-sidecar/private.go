package anvilsidecar

import (
	"os"

	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *AnvilSidecarController) processPod(action sync.SyncAction) error {
	cachedPod := action.GetObject().(corev1.Pod)
	logger := c.logger.WithFields(logrus.Fields{
		"Name":      cachedPod.GetName(),
		"Namespace": cachedPod.GetNamespace(),
	})

	switch action.GetAction() {
	case sync.SyncActionDelete:
		logger.Info("pod was deleted; exiting...")
		os.Exit(0)
	case sync.SyncActionAdd, sync.SyncActionUpdate:
		pod, err := c.podLister.Pods(cachedPod.GetNamespace()).Get(cachedPod.GetName())
		if apierrors.IsNotFound(err) {
			logger.Info("unable to find pod")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting pipeline")
		}

		return c.checkPodContainerStatuses(*pod.DeepCopy(), logger)
	}

	return nil
}

func (c *AnvilSidecarController) checkPodContainerStatuses(original corev1.Pod, logger logrus.FieldLogger) error {
	allContainersHaveExited := true
	exitCode := 0

	for _, status := range original.Status.ContainerStatuses {
		if c.sidecarName != "" && status.Name == c.sidecarName {
			continue
		}

		if status.State.Terminated == nil {
			allContainersHaveExited = false
			continue
		}

		if status.State.Terminated.ExitCode > 0 {
			exitCode = 1
		}
	}

	if allContainersHaveExited {
		logger.Info("firing hook")

		if err := c.fireHook(exitCode, logger); err != nil {
			return errors.Wrap(err, "could not fire hook")
		}

		logger.Infof("hook fired; exiting with code %d", exitCode)
		os.Exit(exitCode)
	} else {
		logger.Info("skipping; work is still being performed")
	}

	return nil
}

func (c *AnvilSidecarController) fireHook(exitCode int, logger logrus.FieldLogger) error {
	// todo: flesh this logic out more
	// todo: add timeout?
	if exitCode == 0 {
		logger.Info("containers were successful")
	} else {
		logger.Info("1 or more containers failed")
	}

	return nil
}
