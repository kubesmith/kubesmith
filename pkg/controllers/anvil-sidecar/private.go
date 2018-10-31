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

	pod, err := c.podLister.Pods(cachedPod.GetNamespace()).Get(cachedPod.GetName())
	if apierrors.IsNotFound(err) {
		logger.Info("unable to find pod")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "error getting pipeline")
	}

	return c.checkPodContainerStatuses(*pod.DeepCopy(), logger)
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
		if exitCode == 0 {
			return c.fireHook(c.processSuccessfulPod, 0, "Successful", original, logger)
		}

		return c.fireHook(c.processFailedPod, exitCode, "Failed", original, logger)
	} else {
		logger.Info("skipping; work is still being performed")
	}

	return nil
}

func (c *AnvilSidecarController) fireHook(
	callback HookCallback,
	exitCode int,
	status string,
	original corev1.Pod,
	logger logrus.FieldLogger,
) error {
	logger = logger.WithField("Status", status)
	logger.Info("firing hook")

	if err := callback(original, logger); err != nil {
		return errors.Wrap(err, "could not fire hook")
	}

	logger.Infof("fired hook; exiting with code %d", exitCode)
	os.Exit(exitCode)

	return nil
}

func (c *AnvilSidecarController) processSuccessfulPod(original corev1.Pod, logger logrus.FieldLogger) error {
	logger.Info("containers were successful")
	return nil
}

func (c *AnvilSidecarController) processFailedPod(original corev1.Pod, logger logrus.FieldLogger) error {
	logger.Info("1 or more containers failed")
	return nil
}
