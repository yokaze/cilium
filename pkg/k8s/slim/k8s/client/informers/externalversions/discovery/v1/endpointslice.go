// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	context "context"
	time "time"

	apidiscoveryv1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/discovery/v1"
	versioned "github.com/cilium/cilium/pkg/k8s/slim/k8s/client/clientset/versioned"
	internalinterfaces "github.com/cilium/cilium/pkg/k8s/slim/k8s/client/informers/externalversions/internalinterfaces"
	discoveryv1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/client/listers/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EndpointSliceInformer provides access to a shared informer and lister for
// EndpointSlices.
type EndpointSliceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() discoveryv1.EndpointSliceLister
}

type endpointSliceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEndpointSliceInformer constructs a new informer for EndpointSlice type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEndpointSliceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEndpointSliceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEndpointSliceInformer constructs a new informer for EndpointSlice type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEndpointSliceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DiscoveryV1().EndpointSlices(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DiscoveryV1().EndpointSlices(namespace).Watch(context.TODO(), options)
			},
		},
		&apidiscoveryv1.EndpointSlice{},
		resyncPeriod,
		indexers,
	)
}

func (f *endpointSliceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEndpointSliceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *endpointSliceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apidiscoveryv1.EndpointSlice{}, f.defaultInformer)
}

func (f *endpointSliceInformer) Lister() discoveryv1.EndpointSliceLister {
	return discoveryv1.NewEndpointSliceLister(f.Informer().GetIndexer())
}
