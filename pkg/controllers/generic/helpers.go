package generic

import "k8s.io/client-go/util/workqueue"

func NewGenericController(name string) *GenericController {
	c := &GenericController{
		Name:  name,
		Queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
	}

	return c
}
