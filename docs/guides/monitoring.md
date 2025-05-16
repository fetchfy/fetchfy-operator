# Monitoring

This guide explains how to monitor your Fetchfy MCP Gateway and the associated MCP services using Prometheus and other observability tools.

## Overview

Fetchfy provides comprehensive observability through:

1. **Prometheus metrics**: For real-time monitoring of the operator and gateway performance
2. **Structured logging**: For debugging and auditing
3. **Kubernetes events**: For tracking important state changes
4. **Gateway status**: For monitoring the health of registered services

## Prometheus Metrics

Fetchfy exposes a variety of Prometheus metrics that provide insights into the gateway's operation.

### Available Metrics

| Metric Name                            | Type      | Description                                    |
| -------------------------------------- | --------- | ---------------------------------------------- |
| `fetchfy_gateway_count`                | Gauge     | Number of MCP gateways managed by the operator |
| `fetchfy_mcp_service_count`            | Gauge     | Number of MCP services by type (tool/agent)    |
| `fetchfy_mcp_request_count`            | Counter   | Number of MCP requests by path and method      |
| `fetchfy_mcp_request_duration_seconds` | Histogram | Duration of MCP requests                       |
| `fetchfy_error_count`                  | Counter   | Number of errors by type                       |

### Accessing Metrics

The metrics are exposed on port 8080 (by default) at the `/metrics` endpoint:

```
http://<operator-pod-ip>:8080/metrics
```

### Prometheus Configuration

To scrape these metrics with Prometheus, add the following to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: "fetchfy-operator"
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - fetchfy-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        action: keep
        regex: fetchfy-operator
      - source_labels: [__meta_kubernetes_pod_container_port_name]
        action: keep
        regex: metrics
```

### ServiceMonitor for Prometheus Operator

If you're using the Prometheus Operator, you can create a ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: fetchfy-operator
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: fetchfy-operator
  endpoints:
    - port: metrics
      interval: 15s
  namespaceSelector:
    matchNames:
      - fetchfy-system
```

## Grafana Dashboards

Fetchfy provides a Grafana dashboard that visualizes the exposed metrics.

### Installing the Dashboard

1. Download the dashboard JSON from the Fetchfy repository:

   ```bash
   curl -L -o fetchfy-dashboard.json https://raw.githubusercontent.com/fetchfy/fetchfy-operator/main/dashboards/operator-dashboard.json
   ```

2. Import the dashboard into Grafana:
   - Navigate to Dashboards > Import
   - Upload the downloaded JSON file
   - Select your Prometheus data source
   - Click Import

### Dashboard Features

The dashboard includes:

- Gateway health status
- MCP service count by type
- Request rate and latency
- Error rate
- Top requested endpoints
- Resource usage

## Structured Logging

Fetchfy uses structured logging to make it easier to parse and analyze logs.

### Accessing Logs

View the operator logs with:

```bash
kubectl logs -n fetchfy-system deployment/fetchfy-operator-controller-manager
```

For more detailed logs, run the operator with increased verbosity:

```yaml
args:
  - --v=5 # Increase verbosity level
```

### Log Levels

Fetchfy uses the following log levels:

- **Error**: Failures that require attention
- **Warning**: Potential issues or degraded functionality
- **Info**: Normal operational events
- **Debug**: Detailed information for troubleshooting (only with increased verbosity)

### Log Fields

Common log fields include:

- `gateway`: The gateway name and namespace
- `service`: The service name and namespace when relevant
- `error`: Error details when applicable
- `component`: Which component generated the log (controller, registry, etc.)

### Example Log Queries

Using tools like Elasticsearch or Loki:

```
# Find registration errors
{namespace="fetchfy-system"} |= "Failed to register service"

# Monitor gateway reconciliation
{namespace="fetchfy-system"} |= "Reconcile" |= "Gateway"

# Track service registration events
{namespace="fetchfy-system"} |= "Registered MCP service"
```

## Kubernetes Events

Fetchfy generates Kubernetes events for significant state changes.

### Viewing Events

```bash
kubectl get events --field-selector involvedObject.kind=Gateway
```

Or for a specific gateway:

```bash
kubectl get events --field-selector involvedObject.kind=Gateway,involvedObject.name=<gateway-name>
```

### Event Types

Common events include:

- **Gateway creation/deletion**: When a Gateway is created or deleted
- **Service registration**: When an MCP service is registered
- **Service deregistration**: When an MCP service is removed
- **Gateway status changes**: When the Gateway transitions between states

## Gateway Status Monitoring

The Gateway custom resource includes a detailed status section that provides real-time information about its state.

### Checking Gateway Status

```bash
kubectl get gateway <gateway-name> -o yaml
```

Look for the `status` section, which includes:

- `address`: The address where the Gateway is available
- `mcpServices`: List of registered MCP services
- `conditions`: Standard Kubernetes conditions reflecting the Gateway's state

### Monitoring with kubectl wait

You can use `kubectl wait` to monitor for specific conditions:

```bash
# Wait for Gateway to be Ready
kubectl wait --for=condition=Ready gateway/<gateway-name> --timeout=60s
```

## Alerting

### Prometheus AlertManager Rules

Example Prometheus alerting rules for monitoring Fetchfy:

```yaml
groups:
  - name: fetchfy
    rules:
      - alert: FetchfyGatewayDown
        expr: sum(up{job="fetchfy-operator"}) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Fetchfy Gateway is down"
          description: "All Fetchfy Gateway instances have been down for 5 minutes."

      - alert: FetchfyHighErrorRate
        expr: rate(fetchfy_error_count[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in Fetchfy Gateway"
          description: "Fetchfy Gateway has a high error rate (> 10%)."

      - alert: FetchfyServiceUnavailable
        expr: sum(fetchfy_mcp_service_count) by (status) > 0 and status == "Unavailable"
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "MCP Service(s) unavailable"
          description: "One or more MCP services are marked as Unavailable."
```

### Health Checks

For health-based monitoring and alerting:

1. **Liveness endpoint**: The operator exposes a liveness endpoint at `:8081/healthz`
2. **Readiness endpoint**: The operator exposes a readiness endpoint at `:8081/readyz`

## Monitoring Best Practices

1. **Dashboard for operator health**: Monitor CPU, memory usage, and restart count
2. **Service monitoring**: Track registered MCP services and their status
3. **Alert on unavailable services**: Get notified when services become unavailable
4. **Track request latency**: Set up thresholds for MCP request duration
5. **Monitor error rates**: Alert on increased error rates
6. **Log aggregation**: Centralize logs for easier debugging
7. **Regular status checks**: Periodically check Gateway status

## Troubleshooting Common Issues

### High Latency

If you notice high latency in MCP requests:

1. Check the service implementation for performance issues
2. Verify network connectivity between components
3. Monitor resource usage on MCP service pods

### Service Registration Issues

If services aren't being registered correctly:

1. Check service labels and annotations
2. Verify gateway selector matches the service labels
3. Check operator logs for registration errors

### High Error Rate

If you see a high error rate:

1. Identify the specific error type from metrics
2. Check operator and service logs for detailed error messages
3. Verify service health and availability

## Extending Monitoring

### Custom Metrics

You can extend monitoring by adding custom metrics to your MCP services:

```go
// Example Go code for a service with custom metrics
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    functionCalls = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "mcp_function_calls_total",
        Help: "Number of function calls",
    }, []string{"function"})

    functionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "mcp_function_duration_seconds",
        Help:    "Duration of function execution in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"function"})
)
```

### Integration with External Systems

Consider integrating with:

- **Logging systems**: Elasticsearch, Loki, Splunk
- **Tracing systems**: Jaeger, Zipkin, OpenTelemetry
- **Visualization**: Custom Grafana dashboards
- **Alerting**: PagerDuty, Slack, OpsGenie
