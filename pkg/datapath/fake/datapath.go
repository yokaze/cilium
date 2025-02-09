// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package fake

import (
	"context"
	"io"

	"github.com/cilium/cilium/pkg/datapath/loader/metrics"
	"github.com/cilium/cilium/pkg/datapath/tunnel"
	datapath "github.com/cilium/cilium/pkg/datapath/types"
	"github.com/cilium/cilium/pkg/testutils/mockmaps"
)

var _ datapath.Datapath = (*FakeDatapath)(nil)

type FakeDatapath struct {
	node           *FakeNodeHandler
	nodeAddressing datapath.NodeAddressing
	loader         datapath.Loader
	lbmap          *mockmaps.LBMockMap
}

// NewDatapath returns a new fake datapath
func NewDatapath() *FakeDatapath {
	return newDatapath(NewNodeAddressing())
}

func newDatapath(na datapath.NodeAddressing) *FakeDatapath {
	return &FakeDatapath{
		node:           NewNodeHandler(),
		nodeAddressing: na,
		loader:         &fakeLoader{},
		lbmap:          mockmaps.NewLBMockMap(),
	}
}

// Node returns a fake handler for node events
func (f *FakeDatapath) Node() datapath.NodeHandler {
	return f.node
}

func (f *FakeDatapath) NodeIDs() datapath.NodeIDHandler {
	return f.node
}

func (f *FakeDatapath) NodeNeighbors() datapath.NodeNeighbors {
	return f.node
}

func (f *FakeDatapath) FakeNode() *FakeNodeHandler {
	return f.node
}

// LocalNodeAddressing returns a fake node addressing implementation of the
// local node
func (f *FakeDatapath) LocalNodeAddressing() datapath.NodeAddressing {
	return f.nodeAddressing
}

// WriteNodeConfig pretends to write the datapath configuration to the writer.
func (f *FakeDatapath) WriteNodeConfig(io.Writer, *datapath.LocalNodeConfiguration) error {
	return nil
}

// WriteNetdevConfig pretends to write the netdev configuration to a writer.
func (f *FakeDatapath) WriteNetdevConfig(io.Writer, datapath.DeviceConfiguration) error {
	return nil
}

// WriteTemplateConfig pretends to write the endpoint configuration to a writer.
func (f *FakeDatapath) WriteTemplateConfig(io.Writer, datapath.EndpointConfiguration) error {
	return nil
}

// WriteEndpointConfig pretends to write the endpoint configuration to a writer.
func (f *FakeDatapath) WriteEndpointConfig(io.Writer, datapath.EndpointConfiguration) error {
	return nil
}

func (f *FakeDatapath) InstallProxyRules(context.Context, uint16, bool, string) error {
	return nil
}

func (f *FakeDatapath) SupportsOriginalSourceAddr() bool {
	return false
}

func (f *FakeDatapath) InstallRules(ctx context.Context, ifName string, quiet, install bool) error {
	return nil
}

func (m *FakeDatapath) GetProxyPorts() map[string]uint16 {
	return nil
}

func (m *FakeDatapath) InstallNoTrackRules(IP string, port uint16, ipv6 bool) error {
	return nil
}

func (m *FakeDatapath) RemoveNoTrackRules(IP string, port uint16, ipv6 bool) error {
	return nil
}

func (f *FakeDatapath) Loader() datapath.Loader {
	return f.loader
}

func (f *FakeDatapath) WireguardAgent() datapath.WireguardAgent {
	return nil
}

func (f *FakeDatapath) Procfs() string {
	return "/proc"
}

func (f *FakeDatapath) LBMap() datapath.LBMap {
	return f.lbmap
}

func (f *FakeDatapath) LBMockMap() *mockmaps.LBMockMap {
	return f.lbmap
}

func (f *FakeDatapath) BandwidthManager() datapath.BandwidthManager {
	return &BandwidthManager{}
}

// Loader is an interface to abstract out loading of datapath programs.
type fakeLoader struct {
}

func (f *fakeLoader) CompileAndLoad(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
	panic("implement me")
}

func (f *fakeLoader) CompileOrLoad(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
	panic("implement me")
}

func (f *fakeLoader) ReloadDatapath(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
	panic("implement me")
}

func (f *fakeLoader) EndpointHash(cfg datapath.EndpointConfiguration) (string, error) {
	panic("implement me")
}

func (f *fakeLoader) Unload(ep datapath.Endpoint) {
}

func (f *fakeLoader) CallsMapPath(id uint16) string {
	return ""
}

func (f *fakeLoader) CustomCallsMapPath(id uint16) string {
	return ""
}

// Reinitialize does nothing.
func (f *fakeLoader) Reinitialize(ctx context.Context, o datapath.BaseProgramOwner, tunnelConfig tunnel.Config, deviceMTU int, iptMgr datapath.IptablesManager, p datapath.Proxy) error {
	return nil
}

func (f *fakeLoader) HostDatapathInitialized() <-chan struct{} {
	return nil
}
