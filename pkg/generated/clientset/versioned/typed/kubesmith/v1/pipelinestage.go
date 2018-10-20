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

package v1

import (
	v1 "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	scheme "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PipelineStagesGetter has a method to return a PipelineStageInterface.
// A group's client should implement this interface.
type PipelineStagesGetter interface {
	PipelineStages(namespace string) PipelineStageInterface
}

// PipelineStageInterface has methods to work with PipelineStage resources.
type PipelineStageInterface interface {
	Create(*v1.PipelineStage) (*v1.PipelineStage, error)
	Update(*v1.PipelineStage) (*v1.PipelineStage, error)
	UpdateStatus(*v1.PipelineStage) (*v1.PipelineStage, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.PipelineStage, error)
	List(opts meta_v1.ListOptions) (*v1.PipelineStageList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PipelineStage, err error)
	PipelineStageExpansion
}

// pipelineStages implements PipelineStageInterface
type pipelineStages struct {
	client rest.Interface
	ns     string
}

// newPipelineStages returns a PipelineStages
func newPipelineStages(c *KubesmithV1Client, namespace string) *pipelineStages {
	return &pipelineStages{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the pipelineStage, and returns the corresponding pipelineStage object, and an error if there is any.
func (c *pipelineStages) Get(name string, options meta_v1.GetOptions) (result *v1.PipelineStage, err error) {
	result = &v1.PipelineStage{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pipelinestages").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PipelineStages that match those selectors.
func (c *pipelineStages) List(opts meta_v1.ListOptions) (result *v1.PipelineStageList, err error) {
	result = &v1.PipelineStageList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pipelinestages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested pipelineStages.
func (c *pipelineStages) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("pipelinestages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a pipelineStage and creates it.  Returns the server's representation of the pipelineStage, and an error, if there is any.
func (c *pipelineStages) Create(pipelineStage *v1.PipelineStage) (result *v1.PipelineStage, err error) {
	result = &v1.PipelineStage{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("pipelinestages").
		Body(pipelineStage).
		Do().
		Into(result)
	return
}

// Update takes the representation of a pipelineStage and updates it. Returns the server's representation of the pipelineStage, and an error, if there is any.
func (c *pipelineStages) Update(pipelineStage *v1.PipelineStage) (result *v1.PipelineStage, err error) {
	result = &v1.PipelineStage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("pipelinestages").
		Name(pipelineStage.Name).
		Body(pipelineStage).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *pipelineStages) UpdateStatus(pipelineStage *v1.PipelineStage) (result *v1.PipelineStage, err error) {
	result = &v1.PipelineStage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("pipelinestages").
		Name(pipelineStage.Name).
		SubResource("status").
		Body(pipelineStage).
		Do().
		Into(result)
	return
}

// Delete takes name of the pipelineStage and deletes it. Returns an error if one occurs.
func (c *pipelineStages) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("pipelinestages").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *pipelineStages) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("pipelinestages").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched pipelineStage.
func (c *pipelineStages) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PipelineStage, err error) {
	result = &v1.PipelineStage{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("pipelinestages").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
