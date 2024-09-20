// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by informer-gen. DO NOT EDIT.

package v2alpha1

import (
	context "context"
	time "time"

	apisciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	versioned "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	internalinterfaces "github.com/cilium/cilium/pkg/k8s/client/informers/externalversions/internalinterfaces"
	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/client/listers/cilium.io/v2alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CiliumPodIPPoolInformer provides access to a shared informer and lister for
// CiliumPodIPPools.
type CiliumPodIPPoolInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() ciliumiov2alpha1.CiliumPodIPPoolLister
}

type ciliumPodIPPoolInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCiliumPodIPPoolInformer constructs a new informer for CiliumPodIPPool type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCiliumPodIPPoolInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCiliumPodIPPoolInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCiliumPodIPPoolInformer constructs a new informer for CiliumPodIPPool type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCiliumPodIPPoolInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CiliumV2alpha1().CiliumPodIPPools().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CiliumV2alpha1().CiliumPodIPPools().Watch(context.TODO(), options)
			},
		},
		&apisciliumiov2alpha1.CiliumPodIPPool{},
		resyncPeriod,
		indexers,
	)
}

func (f *ciliumPodIPPoolInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCiliumPodIPPoolInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *ciliumPodIPPoolInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisciliumiov2alpha1.CiliumPodIPPool{}, f.defaultInformer)
}

func (f *ciliumPodIPPoolInformer) Lister() ciliumiov2alpha1.CiliumPodIPPoolLister {
	return ciliumiov2alpha1.NewCiliumPodIPPoolLister(f.Informer().GetIndexer())
}
