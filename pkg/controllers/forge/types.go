package forge

import (
	"github.com/kubesmith/kubesmith/pkg/controllers/generic"
	api "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/typed/kubesmith/v1"
	listers "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

type ForgeController struct {
	*generic.GenericController

	forgeLister listers.ForgeLister
	forgeClient api.ForgesGetter
	clock       clock.Clock
}
