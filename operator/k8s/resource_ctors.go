// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package k8s

import (
	"errors"
	"fmt"
	"strconv"

	slim_corev1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/cilium/cilium/pkg/hive/cell"
	cilium_api_v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	cilium_api_v2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	"github.com/cilium/cilium/pkg/k8s/client"
	"github.com/cilium/cilium/pkg/k8s/resource"
	"github.com/cilium/cilium/pkg/k8s/utils"
)

func CiliumEndpointResource(lc cell.Lifecycle, cs client.Clientset, opts ...func(*metav1.ListOptions)) (resource.Resource[*cilium_api_v2.CiliumEndpoint], error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	lw := utils.ListerWatcherWithModifiers(
		utils.ListerWatcherFromTyped[*cilium_api_v2.CiliumEndpointList](cs.CiliumV2().CiliumEndpoints("")),
		opts...,
	)
	indexers := cache.Indexers{
		cache.NamespaceIndex:        cache.MetaNamespaceIndexFunc,
		CiliumEndpointIndexIdentity: identityIndexFunc,
	}
	return resource.New[*cilium_api_v2.CiliumEndpoint](
		lc, lw, resource.WithMetric("CiliumEndpoint"), resource.WithIndexers(indexers)), nil
}

func identityIndexFunc(obj interface{}) ([]string, error) {
	switch t := obj.(type) {
	case *cilium_api_v2.CiliumEndpoint:
		if t.Status.Identity != nil {
			id := strconv.FormatInt(t.Status.Identity.ID, 10)
			return []string{id}, nil
		}
		return []string{"0"}, nil
	}
	return nil, fmt.Errorf("%w - found %T", errors.New("object is not a *cilium_api_v2.CiliumEndpoint"), obj)
}

func CiliumEndpointSliceResource(lc cell.Lifecycle, cs client.Clientset, opts ...func(*metav1.ListOptions)) (resource.Resource[*cilium_api_v2alpha1.CiliumEndpointSlice], error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	lw := utils.ListerWatcherWithModifiers(
		utils.ListerWatcherFromTyped[*cilium_api_v2alpha1.CiliumEndpointSliceList](cs.CiliumV2alpha1().CiliumEndpointSlices()),
		opts...,
	)
	return resource.New[*cilium_api_v2alpha1.CiliumEndpointSlice](lc, lw, resource.WithMetric("CiliumEndpointSlice")), nil
}

func PodResource(lc cell.Lifecycle, cs client.Clientset, opts ...func(*metav1.ListOptions)) (resource.Resource[*slim_corev1.Pod], error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	lw := utils.ListerWatcherWithModifiers(
		utils.ListerWatcherFromTyped[*slim_corev1.PodList](cs.Slim().CoreV1().Pods("")),
		opts...,
	)

	indexers := cache.Indexers{
		// Thix index is used for IPAM by the ciliumNodeSynchronizer.
		PodNodeNameIndex: PodNodeNameIndexFunc,
	}

	return resource.New[*slim_corev1.Pod](lc, lw,
			resource.WithMetric("Pod"),
			resource.WithIndexers(indexers),
		),
		nil
}
