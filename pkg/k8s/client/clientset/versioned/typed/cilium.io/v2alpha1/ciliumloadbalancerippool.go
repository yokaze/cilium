// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by client-gen. DO NOT EDIT.

package v2alpha1

import (
	context "context"

	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	scheme "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// CiliumLoadBalancerIPPoolsGetter has a method to return a CiliumLoadBalancerIPPoolInterface.
// A group's client should implement this interface.
type CiliumLoadBalancerIPPoolsGetter interface {
	CiliumLoadBalancerIPPools() CiliumLoadBalancerIPPoolInterface
}

// CiliumLoadBalancerIPPoolInterface has methods to work with CiliumLoadBalancerIPPool resources.
type CiliumLoadBalancerIPPoolInterface interface {
	Create(ctx context.Context, ciliumLoadBalancerIPPool *ciliumiov2alpha1.CiliumLoadBalancerIPPool, opts v1.CreateOptions) (*ciliumiov2alpha1.CiliumLoadBalancerIPPool, error)
	Update(ctx context.Context, ciliumLoadBalancerIPPool *ciliumiov2alpha1.CiliumLoadBalancerIPPool, opts v1.UpdateOptions) (*ciliumiov2alpha1.CiliumLoadBalancerIPPool, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, ciliumLoadBalancerIPPool *ciliumiov2alpha1.CiliumLoadBalancerIPPool, opts v1.UpdateOptions) (*ciliumiov2alpha1.CiliumLoadBalancerIPPool, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*ciliumiov2alpha1.CiliumLoadBalancerIPPool, error)
	List(ctx context.Context, opts v1.ListOptions) (*ciliumiov2alpha1.CiliumLoadBalancerIPPoolList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *ciliumiov2alpha1.CiliumLoadBalancerIPPool, err error)
	CiliumLoadBalancerIPPoolExpansion
}

// ciliumLoadBalancerIPPools implements CiliumLoadBalancerIPPoolInterface
type ciliumLoadBalancerIPPools struct {
	*gentype.ClientWithList[*ciliumiov2alpha1.CiliumLoadBalancerIPPool, *ciliumiov2alpha1.CiliumLoadBalancerIPPoolList]
}

// newCiliumLoadBalancerIPPools returns a CiliumLoadBalancerIPPools
func newCiliumLoadBalancerIPPools(c *CiliumV2alpha1Client) *ciliumLoadBalancerIPPools {
	return &ciliumLoadBalancerIPPools{
		gentype.NewClientWithList[*ciliumiov2alpha1.CiliumLoadBalancerIPPool, *ciliumiov2alpha1.CiliumLoadBalancerIPPoolList](
			"ciliumloadbalancerippools",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *ciliumiov2alpha1.CiliumLoadBalancerIPPool { return &ciliumiov2alpha1.CiliumLoadBalancerIPPool{} },
			func() *ciliumiov2alpha1.CiliumLoadBalancerIPPoolList {
				return &ciliumiov2alpha1.CiliumLoadBalancerIPPoolList{}
			},
		),
	}
}
