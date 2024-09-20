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

// CiliumBGPPeerConfigInformer provides access to a shared informer and lister for
// CiliumBGPPeerConfigs.
type CiliumBGPPeerConfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() ciliumiov2alpha1.CiliumBGPPeerConfigLister
}

type ciliumBGPPeerConfigInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCiliumBGPPeerConfigInformer constructs a new informer for CiliumBGPPeerConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCiliumBGPPeerConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCiliumBGPPeerConfigInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCiliumBGPPeerConfigInformer constructs a new informer for CiliumBGPPeerConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCiliumBGPPeerConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CiliumV2alpha1().CiliumBGPPeerConfigs().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CiliumV2alpha1().CiliumBGPPeerConfigs().Watch(context.TODO(), options)
			},
		},
		&apisciliumiov2alpha1.CiliumBGPPeerConfig{},
		resyncPeriod,
		indexers,
	)
}

func (f *ciliumBGPPeerConfigInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCiliumBGPPeerConfigInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *ciliumBGPPeerConfigInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisciliumiov2alpha1.CiliumBGPPeerConfig{}, f.defaultInformer)
}

func (f *ciliumBGPPeerConfigInformer) Lister() ciliumiov2alpha1.CiliumBGPPeerConfigLister {
	return ciliumiov2alpha1.NewCiliumBGPPeerConfigLister(f.Informer().GetIndexer())
}
