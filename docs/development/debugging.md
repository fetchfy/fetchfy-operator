---
title: Debugging Guide
description: Advanced techniques for debugging the Fetchfy MCP Gateway Operator
---

# Debugging Guide

This guide provides strategies and techniques for debugging the Fetchfy MCP Gateway Operator during development and in production environments.

## Local Debugging

### Using Delve

[Delve](https://github.com/go-delve/delve) is a powerful debugger for Go applications. To use it with the Fetchfy Operator:

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run the operator with Delve
dlv debug ./cmd/main.go -- --metrics-bind-address=:8080
```

This starts the debugger and allows you to:

- Set breakpoints
- Inspect variables
- Step through code execution
- Evaluate expressions

### VS Code Integration

If you're using VS Code, you can configure launch settings for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Operator",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/main.go",
      "args": ["--metrics-bind-address=:8080"],
      "env": {
        "KUBECONFIG": "${env:HOME}/.kube/config",
        "WATCH_NAMESPACE": ""
      }
    }
  ]
}
```

Save this configuration in `.vscode/launch.json` and use the VS Code debugging tools.

## Enhanced Logging

### Enabling Debug Logs

To enable verbose logging during development:

```bash
# When running locally
make run ARGS="--zap-log-level=debug"

# Or in a deployed environment
kubectl edit deployment -n fetchfy-system fetchfy-controller-manager
```

In the deployment, add or modify the args:

```yaml
spec:
  template:
    spec:
      containers:
        - name: manager
          args:
            - "--zap-log-level=debug"
```

### Structured Logging

The operator uses structured logging with [zap](https://github.com/uber-go/zap). To add contextual information:

```go
import "sigs.k8s.io/controller-runtime/pkg/log"

func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx).
        WithValues("gateway", req.NamespacedName)

    logger.Info("Reconciling Gateway")

    // Later in the code
    logger.V(1).Info("Processing service",
        "serviceName", service.Name,
        "namespace", service.Namespace)
}
```

## Inspecting Kubernetes Resources

### Checking Gateway Resources

View gateway resources and their status:

```bash
kubectl get gateways -A
kubectl describe gateway <gateway-name> -n <namespace>
```

### Inspecting Controller Logs

```bash
kubectl logs -n fetchfy-system -l control-plane=controller-manager -c manager --tail=100 -f
```

Add grep to filter specific events:

```bash
kubectl logs -n fetchfy-system -l control-plane=controller-manager -c manager -f | grep "Reconciling Gateway"
```

### Examining Events

Kubernetes events provide valuable debugging information:

```bash
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

Filter events related to your resource:

```bash
kubectl get events -n <namespace> --field-selector involvedObject.name=<gateway-name>
```

## Investigating MCP Communication

### Testing MCP Server Connectivity

```bash
# Port-forward the MCP Gateway service
kubectl port-forward -n fetchfy-system svc/<gateway-service> 8080:8080

# Test with curl
curl http://localhost:8080/v1/tools -H "Content-Type: application/json"
```

### Analyzing Network Issues

Use tcpdump to capture and analyze traffic:

```bash
# On the gateway pod
kubectl exec -it -n fetchfy-system <gateway-pod> -- tcpdump -n port 8080

# Save to a file for analysis
kubectl exec -it -n fetchfy-system <gateway-pod> -- tcpdump -n -w /tmp/dump.pcap port 8080
kubectl cp fetchfy-system/<gateway-pod>:/tmp/dump.pcap ./dump.pcap
```

## Debugging Reconciliation Issues

### Common Reconciliation Problems

1. **Resource not updating**: Check finalizers or validation issues
2. **Controller crashing**: Look for panics in the logs
3. **Permissions issues**: Verify RBAC settings

### Inspecting Controller Cache

To verify what the controller sees in its cache:

```go
import (
    "context"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func debugCache(c client.Client) {
    var serviceList corev1.ServiceList
    err := c.List(context.Background(), &serviceList,
        client.MatchingLabels{"mcp.fetchfy.io/type": "tool"})
    if err != nil {
        log.Error(err, "Failed to list services")
        return
    }

    for _, svc := range serviceList.Items {
        log.Info("Service in cache",
            "name", svc.Name,
            "namespace", svc.Namespace)
    }
}
```

## Metrics and Monitoring

### Prometheus Metrics

The operator exposes metrics that can help with debugging:

```bash
# Port-forward the metrics endpoint
kubectl port-forward -n fetchfy-system <controller-pod> 8080:8080

# Query metrics
curl http://localhost:8080/metrics
```

Key metrics to look for:

- `controller_runtime_reconcile_total`: Total number of reconciliations
- `controller_runtime_reconcile_errors_total`: Total number of reconciliation errors
- `controller_runtime_reconcile_time_seconds`: Time taken for reconciliations
- `fetchfy_registered_services`: Number of registered services (custom metric)

### Adding Custom Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    registeredServices = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "fetchfy_registered_services",
            Help: "Number of services registered with the MCP Gateway",
        },
        []string{"gateway", "namespace", "type"},
    )
)

func init() {
    metrics.Registry.MustRegister(registeredServices)
}
```

## Profiling

### Runtime Profiling

Enable Go's built-in profiling endpoints:

```bash
# When running locally
make run ARGS="--profiler-address=:6060"
```

Access profiles using:

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Analyzing Performance Bottlenecks

To identify slow reconciliation loops:

```go
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    start := time.Now()
    defer func() {
        log.Info("Reconciliation completed",
            "gateway", req.NamespacedName,
            "duration", time.Since(start))
    }()

    // Reconciliation logic
}
```

## Production Debugging Techniques

### Using kubectl-debug

[kubectl-debug](https://github.com/aylei/kubectl-debug) helps debugging in production:

```bash
# Install kubectl-debug
kubectl krew install debug

# Debug a running pod
kubectl debug -n fetchfy-system <pod-name> --image=busybox
```

### Core Dumps

Configure the operator to generate core dumps:

```yaml
spec:
  template:
    spec:
      containers:
        - name: manager
          env:
            - name: GOTRACEBACK
              value: "crash"
          volumeMounts:
            - name: core-dumps
              mountPath: /tmp
      volumes:
        - name: core-dumps
          emptyDir: {}
```

### Running with Race Detection

During development, enable race detection:

```bash
# Build with race detection
go build -race -o bin/manager cmd/main.go

# Or using make
make build ARGS="-race"
```

## Next Steps

- Review the [Development Setup Guide](./setup.md) for environment configuration
- Check the [Contributing Guide](./contributing.md) for code contribution guidelines
- Explore the API details in the [Gateway CRD Reference](../api-reference/gateway-crd.md)
