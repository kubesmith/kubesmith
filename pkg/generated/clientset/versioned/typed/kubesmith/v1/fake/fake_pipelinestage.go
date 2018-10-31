/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	kubesmithv1 "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePipelineStages implements PipelineStageInterface
type FakePipelineStages struct {
	Fake *FakeKubesmithV1
	ns   string
}

var pipelinestagesResource = schema.GroupVersionResource{Group: "kubesmith.io", Version: "v1", Resource: "pipelinestages"}

var pipelinestagesKind = schema.GroupVersionKind{Group: "kubesmith.io", Version: "v1", Kind: "PipelineStage"}

// Get takes name of the pipelineStage, and returns the corresponding pipelineStage object, and an error if there is any.
func (c *FakePipelineStages) Get(name string, options v1.GetOptions) (result *kubesmithv1.PipelineStage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pipelinestagesResource, c.ns, name), &kubesmithv1.PipelineStage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubesmithv1.PipelineStage), err
}

// List takes label and field selectors, and returns the list of PipelineStages that match those selectors.
func (c *FakePipelineStages) List(opts v1.ListOptions) (result *kubesmithv1.PipelineStageList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pipelinestagesResource, pipelinestagesKind, c.ns, opts), &kubesmithv1.PipelineStageList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubesmithv1.PipelineStageList{ListMeta: obj.(*kubesmithv1.PipelineStageList).ListMeta}
	for _, item := range obj.(*kubesmithv1.PipelineStageList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pipelineStages.
func (c *FakePipelineStages) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pipelinestagesResource, c.ns, opts))

}

// Create takes the representation of a pipelineStage and creates it.  Returns the server's representation of the pipelineStage, and an error, if there is any.
func (c *FakePipelineStages) Create(pipelineStage *kubesmithv1.PipelineStage) (result *kubesmithv1.PipelineStage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pipelinestagesResource, c.ns, pipelineStage), &kubesmithv1.PipelineStage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubesmithv1.PipelineStage), err
}

// Update takes the representation of a pipelineStage and updates it. Returns the server's representation of the pipelineStage, and an error, if there is any.
func (c *FakePipelineStages) Update(pipelineStage *kubesmithv1.PipelineStage) (result *kubesmithv1.PipelineStage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pipelinestagesResource, c.ns, pipelineStage), &kubesmithv1.PipelineStage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubesmithv1.PipelineStage), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePipelineStages) UpdateStatus(pipelineStage *kubesmithv1.PipelineStage) (*kubesmithv1.PipelineStage, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pipelinestagesResource, "status", c.ns, pipelineStage), &kubesmithv1.PipelineStage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubesmithv1.PipelineStage), err
}

// Delete takes name of the pipelineStage and deletes it. Returns an error if one occurs.
func (c *FakePipelineStages) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(pipelinestagesResource, c.ns, name), &kubesmithv1.PipelineStage{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePipelineStages) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pipelinestagesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubesmithv1.PipelineStageList{})
	return err
}

// Patch applies the patch and returns the patched pipelineStage.
func (c *FakePipelineStages) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubesmithv1.PipelineStage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(pipelinestagesResource, c.ns, name, data, subresources...), &kubesmithv1.PipelineStage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubesmithv1.PipelineStage), err
}
