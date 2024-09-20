// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/typed/cilium.io/v2alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeCiliumBGPNodeConfigs implements CiliumBGPNodeConfigInterface
type fakeCiliumBGPNodeConfigs struct {
	*gentype.FakeClientWithList[*v2alpha1.CiliumBGPNodeConfig, *v2alpha1.CiliumBGPNodeConfigList]
	Fake *FakeCiliumV2alpha1
}

func newFakeCiliumBGPNodeConfigs(fake *FakeCiliumV2alpha1) ciliumiov2alpha1.CiliumBGPNodeConfigInterface {
	return &fakeCiliumBGPNodeConfigs{
		gentype.NewFakeClientWithList[*v2alpha1.CiliumBGPNodeConfig, *v2alpha1.CiliumBGPNodeConfigList](
			fake.Fake,
			"",
			v2alpha1.SchemeGroupVersion.WithResource("ciliumbgpnodeconfigs"),
			v2alpha1.SchemeGroupVersion.WithKind("CiliumBGPNodeConfig"),
			func() *v2alpha1.CiliumBGPNodeConfig { return &v2alpha1.CiliumBGPNodeConfig{} },
			func() *v2alpha1.CiliumBGPNodeConfigList { return &v2alpha1.CiliumBGPNodeConfigList{} },
			func(dst, src *v2alpha1.CiliumBGPNodeConfigList) { dst.ListMeta = src.ListMeta },
			func(list *v2alpha1.CiliumBGPNodeConfigList) []*v2alpha1.CiliumBGPNodeConfig {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v2alpha1.CiliumBGPNodeConfigList, items []*v2alpha1.CiliumBGPNodeConfig) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
