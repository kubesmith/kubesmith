package forge

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	"github.com/kubesmith/kubesmith/pkg/sync"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *ForgeController) processForge(action sync.SyncAction) error {
	cachedForge := action.GetObject().(api.Forge)
	logger := c.logger.WithFields(logrus.Fields{
		"Name": cachedForge.GetName(),
	})

	switch action.GetAction() {
	case sync.SyncActionDelete:
		if err := c.processDeletedForge(*cachedForge.DeepCopy(), logger); err != nil {
			return err
		}
	default:
		forge, err := c.forgeLister.Forges(cachedForge.GetNamespace()).Get(cachedForge.GetName())
		if apierrors.IsNotFound(err) {
			c.logger.Info("unable to find forge")
			return nil
		} else if err != nil {
			return errors.Wrap(err, "error getting forge")
		}

		_ = forge
		logger.Info("todo: processing forge")
	}

	return nil
}

func (c *ForgeController) processDeletedForge(original api.Forge, logger logrus.FieldLogger) error {
	logger.Info("todo: processing deleted forge")
	return nil
}

func (c *ForgeController) patchForge(updated, original api.Forge) (*api.Forge, error) {
	patchType, patchBytes, err := updated.GetPatchFromOriginal(original)
	if err != nil {
		return nil, err
	}

	return c.kubesmithClient.Forges(original.GetNamespace()).Patch(original.GetName(), patchType, patchBytes)
}
