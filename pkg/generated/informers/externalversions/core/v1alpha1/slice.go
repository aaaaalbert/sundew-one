/*
Copyright 2023 Contributors to the EdgeNet project.

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

package v1alpha1

import (
	"context"
	time "time"

	corev1alpha1 "github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha1"
	versioned "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/EdgeNet-project/edgenet/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/EdgeNet-project/edgenet/pkg/generated/listers/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SliceInformer provides access to a shared informer and lister for
// Slices.
type SliceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SliceLister
}

type sliceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewSliceInformer constructs a new informer for Slice type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSliceInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSliceInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredSliceInformer constructs a new informer for Slice type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSliceInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1alpha1().Slices().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1alpha1().Slices().Watch(context.TODO(), options)
			},
		},
		&corev1alpha1.Slice{},
		resyncPeriod,
		indexers,
	)
}

func (f *sliceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSliceInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *sliceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&corev1alpha1.Slice{}, f.defaultInformer)
}

func (f *sliceInformer) Lister() v1alpha1.SliceLister {
	return v1alpha1.NewSliceLister(f.Informer().GetIndexer())
}
