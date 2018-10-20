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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	kubesmith_v1 "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
	versioned "github.com/kubesmith/kubesmith/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/kubesmith/kubesmith/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/kubesmith/kubesmith/pkg/generated/listers/kubesmith/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// PipelineJobInformer provides access to a shared informer and lister for
// PipelineJobs.
type PipelineJobInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.PipelineJobLister
}

type pipelineJobInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPipelineJobInformer constructs a new informer for PipelineJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPipelineJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPipelineJobInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPipelineJobInformer constructs a new informer for PipelineJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPipelineJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubesmithV1().PipelineJobs(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubesmithV1().PipelineJobs(namespace).Watch(options)
			},
		},
		&kubesmith_v1.PipelineJob{},
		resyncPeriod,
		indexers,
	)
}

func (f *pipelineJobInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPipelineJobInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *pipelineJobInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubesmith_v1.PipelineJob{}, f.defaultInformer)
}

func (f *pipelineJobInformer) Lister() v1.PipelineJobLister {
	return v1.NewPipelineJobLister(f.Informer().GetIndexer())
}
