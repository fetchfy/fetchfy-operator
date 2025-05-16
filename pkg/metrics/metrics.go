/*
Copyright 2025.

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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// GatewayCount tracks the number of MCP gateways
	GatewayCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fetchfy_gateway_count",
			Help: "Number of MCP gateways managed by the operator",
		},
	)

	// ServiceCount tracks the number of MCP services
	ServiceCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fetchfy_mcp_service_count",
			Help: "Number of MCP services by type",
		},
		[]string{"type"},
	)

	// RequestCount tracks the number of MCP requests
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fetchfy_mcp_request_count",
			Help: "Number of MCP requests by path",
		},
		[]string{"path", "method"},
	)

	// RequestDuration tracks the duration of MCP requests
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fetchfy_mcp_request_duration_seconds",
			Help:    "Duration of MCP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// ErrorCount tracks the number of errors
	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fetchfy_error_count",
			Help: "Number of errors by type",
		},
		[]string{"type"},
	)
)

// init registers all the metrics with the controller-runtime metrics registry
func init() {
	// Register all metrics with the controller-runtime metrics registry
	metrics.Registry.MustRegister(
		GatewayCount,
		ServiceCount,
		RequestCount,
		RequestDuration,
		ErrorCount,
	)
}
