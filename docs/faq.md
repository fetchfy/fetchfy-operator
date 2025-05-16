---
title: Frequently Asked Questions
description: Answers to common questions about the Fetchfy MCP Gateway Operator
---

# Frequently Asked Questions

This page provides answers to commonly asked questions about the Fetchfy MCP Gateway Operator.

## General Questions

### What is the Fetchfy MCP Gateway Operator?

The Fetchfy MCP Gateway Operator is a Kubernetes operator that simplifies the integration of AI model tools and agents in Kubernetes environments by implementing the Model Context Protocol (MCP). It automatically discovers and registers MCP-compatible services, making them available for use by AI models through a unified gateway.

### What is the Model Context Protocol (MCP)?

The Model Context Protocol (MCP) is a standardized communication protocol designed for interaction between AI models, tools, and agents. It enables dynamic tool discovery, parameter validation, structured communication, and runtime integration.

### What are the key components of the Fetchfy Operator?

Key components include:

- Gateway Custom Resource for configuration
- MCP Server for handling protocol requests
- Service Registry for tracking available services
- Service Watcher for Kubernetes service monitoring

### What's the difference between MCP Tools and Agents?

- **Tools**: Simple, stateless services that perform specific tasks with well-defined inputs and outputs.
- **Agents**: More complex, stateful services that can perform multi-step reasoning, maintain context across interactions, and execute complex workflows.

## Installation and Setup

### What are the prerequisites for installing the Fetchfy Operator?

Prerequisites include:

- Kubernetes cluster (v1.19+)
- Kubectl configured to access your cluster
- Helm (optional, for Helm-based installation)
- Administrative access to the cluster for CRD installation

### How do I install the Fetchfy Operator?

You can install the Fetchfy Operator using Helm:

```bash
helm repo add fetchfy https://fetchfy.github.io/charts
helm repo update
helm install fetchfy-operator fetchfy/fetchfy-operator -n fetchfy-system --create-namespace
```

Or using kubectl with manifests:

```bash
kubectl apply -f https://github.com/fetchfy/fetchfy-operator/releases/download/v0.1.0/fetchfy-operator.yaml
```

### How do I verify the operator is working correctly?

Check that the operator pod is running:

```bash
kubectl get pods -n fetchfy-system
```

And that the CRDs are installed:

```bash
kubectl get crds | grep fetchfy.io
```

## Configuration

### How do I configure a Gateway?

Create a Gateway resource in your desired namespace:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: mcp-gateway
  namespace: your-namespace
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
```

### Can I run multiple Gateways?

Yes, you can create multiple Gateway resources in different namespaces or even in the same namespace. Each Gateway will manage its own set of services based on its `serviceSelector`.

### How do I enable TLS for secure communication?

Add TLS configuration to your Gateway resource:

```yaml
spec:
  tls:
    enabled: true
    secretName: gateway-tls-cert
```

And create a Kubernetes Secret with TLS certificate and key.

## Service Integration

### How do I make my service discoverable by the Gateway?

Add the appropriate label to your Kubernetes Service:

```yaml
metadata:
  labels:
    mcp.fetchfy.io/type: tool # or "agent" for agents
```

### What path should my service expose for MCP discovery?

By default, the Gateway looks for `/mcp/definition`. If your service exposes the MCP definition at a different path, add an annotation:

```yaml
metadata:
  annotations:
    mcp.fetchfy.io/path: "/api/mcp/definition"
```

### How do I know if my service has been registered?

Check the Gateway status:

```bash
kubectl describe gateway mcp-gateway -n your-namespace
```

Look for your service in the `status.registeredServices` section.

## Troubleshooting

### My service isn't being discovered by the Gateway

Check the following:

1. Ensure your service has the correct labels (`mcp.fetchfy.io/type: tool|agent`)
2. Verify the Gateway's `serviceSelector` matches your service's labels
3. Check that your service exposes the MCP definition endpoint correctly
4. Look at the operator logs for any errors:

```bash
kubectl logs -n fetchfy-system deployment/fetchfy-controller-manager -c manager
```

### I'm getting TLS errors when connecting to the Gateway

Common TLS issues:

1. Incorrect certificate or key in the Secret
2. Certificate not valid for the Gateway's service name
3. Client not trusting the certificate

Check the Gateway logs:

```bash
kubectl logs -n your-namespace deployment/mcp-gateway
```

### The operator is crashing or restarting frequently

Check the operator logs:

```bash
kubectl logs -n fetchfy-system deployment/fetchfy-controller-manager -c manager
```

Verify the operator has sufficient resources and appropriate RBAC permissions.

## Performance and Scaling

### How many services can a single Gateway handle?

A single Gateway can typically handle hundreds of services, depending on:

- Available resources (CPU, memory)
- Frequency of tool/agent invocations
- Complexity of registered services

### How do I scale the Gateway for production use?

For production, consider:

1. Setting appropriate resource requests and limits
2. Running multiple Gateway replicas
3. Using horizontal pod autoscaling
4. Implementing service mesh for advanced routing

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mcp-gateway-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mcp-gateway
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### How can I monitor Gateway performance?

The Gateway exposes Prometheus metrics that you can scrape with Prometheus and visualize with Grafana. Key metrics include:

- Request latency
- Error rates
- Number of registered services
- Service health status

## Development and Contribution

### How can I contribute to the Fetchfy Operator?

1. Fork the [GitHub repository](https://github.com/fetchfy/fetchfy-operator)
2. Follow the setup instructions in the [Development Setup Guide](../development/setup.md)
3. Make your changes and submit a pull request
4. Follow the guidelines in the [Contributing Guide](../development/contributing.md)

### How do I report bugs or request features?

You can:

1. Open an issue on [GitHub](https://github.com/fetchfy/fetchfy-operator/issues)
2. Join our community discussions
3. Contact the maintainers directly

### How can I extend the Operator's functionality?

The Fetchfy Operator is designed to be extensible. You can:

1. Add custom controllers for new resource types
2. Implement plugins for additional functionality
3. Contribute to the core codebase

See the [Development Guide](../development/setup.md) for more information.
