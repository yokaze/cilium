// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by lister-gen. DO NOT EDIT.

package v2alpha1

import (
	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// CiliumBGPPeerConfigLister helps list CiliumBGPPeerConfigs.
// All objects returned here must be treated as read-only.
type CiliumBGPPeerConfigLister interface {
	// List lists all CiliumBGPPeerConfigs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*ciliumiov2alpha1.CiliumBGPPeerConfig, err error)
	// Get retrieves the CiliumBGPPeerConfig from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*ciliumiov2alpha1.CiliumBGPPeerConfig, error)
	CiliumBGPPeerConfigListerExpansion
}

// ciliumBGPPeerConfigLister implements the CiliumBGPPeerConfigLister interface.
type ciliumBGPPeerConfigLister struct {
	listers.ResourceIndexer[*ciliumiov2alpha1.CiliumBGPPeerConfig]
}

// NewCiliumBGPPeerConfigLister returns a new CiliumBGPPeerConfigLister.
func NewCiliumBGPPeerConfigLister(indexer cache.Indexer) CiliumBGPPeerConfigLister {
	return &ciliumBGPPeerConfigLister{listers.New[*ciliumiov2alpha1.CiliumBGPPeerConfig](indexer, ciliumiov2alpha1.Resource("ciliumbgppeerconfig"))}
}
