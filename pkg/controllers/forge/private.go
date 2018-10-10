package forge

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func (c *ForgeController) processForge(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrap(err, "error splitting queue key")
	}

	forge, err := c.forgeLister.Forges(ns).Get(name)
	if err != nil {
		return errors.Wrap(err, "error getting forge")
	}

	spew.Dump(forge)

	return nil
}

func (c *ForgeController) resync() {
	list, err := c.forgeLister.List(labels.Everything())
	if err != nil {
		glog.V(1).Info("error listing forges")
		glog.Error(err)
		return
	}

	for _, forge := range list {
		key, err := cache.MetaNamespaceKeyFunc(forge)
		if err != nil {
			glog.Errorf("error generating key for forge; key: %s", forge.Name)
			continue
		}

		c.Queue.Add(key)
	}
}