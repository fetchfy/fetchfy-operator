---
title: Configuration Guide
description: Learn how to configure the Fetchfy MCP Gateway Operator
---

# Configuration Guide

This guide provides detailed information on how to configure the Fetchfy MCP Gateway Operator in your Kubernetes cluster.

## Gateway CRD Configuration

The Gateway Custom Resource allows you to configure how the MCP Gateway Operator operates. Below are the key configuration options:

### Basic Configuration

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: mcp-gateway
spec:
  # The port on which the MCP server will listen
  mcpPort: 8080

  # Service selector to identify which services should be registered with this gateway
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
```

### Advanced Configuration Options

#### TLS Configuration

To enable secure communication using TLS:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: secure-mcp-gateway
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  tls:
    enabled: true
    secretName: mcp-gateway-tls
    # Optional certificate rotation interval in hours (default: 24)
    rotationInterval: 48
```

The Secret referenced by `secretName` should contain `tls.crt` and `tls.key` entries.

#### Gateway Logging

Configure logging levels for the gateway:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: mcp-gateway
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  logging:
    level: info # Available options: debug, info, warn, error
    format: json # Available options: json, text
```

#### Connection Settings

Configure connection-related parameters:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: mcp-gateway
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  connectionSettings:
    maxConnections: 1000
    connectionTimeout: 30 # in seconds
    keepAliveInterval: 60 # in seconds
```

## Environment Variables

The Fetchfy Operator supports the following environment variables:

| Variable                    | Description                      | Default Value      |
| --------------------------- | -------------------------------- | ------------------ |
| `LOG_LEVEL`                 | Sets the log level               | `info`             |
| `METRICS_PORT`              | Port for Prometheus metrics      | `8080`             |
| `LEADER_ELECTION_NAMESPACE` | Namespace for leader election    | Operator namespace |
| `WATCH_NAMESPACE`           | Namespace to watch for resources | All namespaces     |

## Kubernetes Pod Configuration

### Resource Requests and Limits

It's recommended to set appropriate resource requests and limits for the operator:

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

### Pod Security Context

To enhance security, configure the security context for the operator:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

## Configuring High Availability

For production deployments, it's recommended to run multiple replicas of the operator for high availability:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fetchfy-controller-manager
spec:
  replicas: 3
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      # ... other configuration ...
```

The operator uses Kubernetes leader election to ensure only one instance is actively reconciling resources.

## Next Steps

After configuring the Fetchfy MCP Gateway Operator, proceed to:

- [Creating a Gateway](../guides/creating-gateway.md) to set up your first gateway instance
- [Deploying MCP Services](../guides/deploying-services.md) to register tools and agents with your gateway
- [Security Guide](../guides/security.md) for securing your MCP communication
