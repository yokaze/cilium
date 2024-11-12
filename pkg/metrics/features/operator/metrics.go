// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package features

import (
	"github.com/cilium/cilium/operator/option"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/metrics/metric"
)

type Metrics struct {
	ACLBGatewayAPIEnabled        metric.Gauge
	ACLBIngressControllerEnabled metric.Gauge
}

const (
	subsystemACLB = "feature_adv_connect_and_lb"
)

// NewMetrics returns all feature metrics. If 'withDefaults' is set, then
// all metrics will have defined all of their possible values.
func NewMetrics(withDefaults bool) Metrics {
	return Metrics{
		ACLBGatewayAPIEnabled: metric.NewGauge(metric.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystemACLB,
			Help:      "GatewayAPI enabled on the operator",
			Name:      "gateway_api_enabled",
		}),

		ACLBIngressControllerEnabled: metric.NewGauge(metric.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystemACLB,
			Help:      "IngressController enabled on the operator",
			Name:      "ingress_controller_enabled",
		}),
	}
}

type featureMetrics interface {
	update(params enabledFeatures, config *option.OperatorConfig)
}

func (m Metrics) update(params enabledFeatures, config *option.OperatorConfig) {
	if config.EnableGatewayAPI {
		m.ACLBGatewayAPIEnabled.Add(1)
	}
	if params.IsIngressControllerEnabled() {
		m.ACLBIngressControllerEnabled.Add(1)
	}
}
