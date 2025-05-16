---
title: Basic Gateway Example
description: A simple example of setting up a basic Fetchfy MCP Gateway
---

# Basic Gateway Example

This example demonstrates how to set up a basic MCP Gateway using the Fetchfy Operator. We'll walk through the process of deploying the gateway and verifying it works correctly.

## Prerequisites

Before starting, ensure:

- The Fetchfy Operator is installed in your cluster
- You have kubectl configured to access your cluster
- You have permissions to create resources in your target namespace

## Step 1: Create a Namespace

First, let's create a dedicated namespace for our gateway:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mcp-example
```

Save this as `namespace.yaml` and apply it:

```bash
kubectl apply -f namespace.yaml
```

## Step 2: Define the Gateway

Create a basic MCP Gateway configuration:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: basic-gateway
  namespace: mcp-example
spec:
  # Port on which the MCP server will listen
  mcpPort: 8080

  # Select services with the mcp-tool label
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
```

Save this as `basic-gateway.yaml` and apply it:

```bash
kubectl apply -f basic-gateway.yaml
```

## Step 3: Verify Gateway Deployment

Check that the gateway has been deployed successfully:

```bash
kubectl get gateway -n mcp-example
```

You should see output similar to:

```
NAME            STATUS    AGE    SERVICES
basic-gateway   Running   45s    0
```

To see more details about the gateway:

```bash
kubectl describe gateway basic-gateway -n mcp-example
```

## Step 4: Access the Gateway

To access the MCP Gateway from within the cluster, services can use:

```
basic-gateway.mcp-example.svc:8080
```

For testing purposes, you can port-forward the gateway service to your local machine:

```bash
kubectl port-forward -n mcp-example svc/basic-gateway 8080:8080
```

Then access it at `http://localhost:8080`.

## Step 5: Test the Gateway API

With the port-forward active, you can test the gateway API:

```bash
# List available tools
curl http://localhost:8080/v1/tools

# Check gateway status
curl http://localhost:8080/v1/status
```

The responses should be empty or minimal since we haven't registered any tools yet.

## Advanced Configuration

### Enabling TLS

To enable TLS for secure communication:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: secure-gateway
  namespace: mcp-example
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  tls:
    enabled: true
    secretName: gateway-tls-cert
```

You'll need to create a TLS secret:

```bash
# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt -subj "/CN=basic-gateway.mcp-example.svc"

# Create Kubernetes secret
kubectl create secret tls gateway-tls-cert \
  --key tls.key \
  --cert tls.crt \
  -n mcp-example
```

### Setting Resource Limits

For production use, set resource limits:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: basic-gateway
  namespace: mcp-example
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  resources:
    requests:
      memory: "64Mi"
      cpu: "100m"
    limits:
      memory: "128Mi"
      cpu: "200m"
```

## Common Issues and Troubleshooting

### Gateway Not Starting

If the gateway doesn't start, check the operator logs:

```bash
kubectl logs -n fetchfy-system -l app=fetchfy-controller-manager -c manager
```

### Cannot Connect to Gateway

If you can't connect to the gateway:

1. Verify the gateway is running:

```bash
kubectl get pods -n mcp-example -l app=basic-gateway
```

2. Check gateway logs:

```bash
kubectl logs -n mcp-example -l app=basic-gateway
```

3. Ensure network policies allow access to the gateway.

### No Services Registered

If no services are showing up in the gateway:

1. Verify your services have the correct labels:

```bash
kubectl get svc -A -l mcp.fetchfy.io/type=tool
```

2. Check if the label selector in the gateway matches your services.

## Next Steps

Now that you have a basic gateway set up, you can:

- [Deploy MCP Tools](./mcp-tool.md) to register them with the gateway
- [Deploy MCP Agents](./mcp-agent.md) for more complex AI capabilities
- Configure [monitoring](../guides/monitoring.md) for your gateway

## Complete Example

For convenience, here's a complete example you can apply:

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: mcp-example
---
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: basic-gateway
  namespace: mcp-example
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  logging:
    level: info
    format: json
```

Save this as `complete-gateway.yaml` and apply it:

```bash
kubectl apply -f complete-gateway.yaml
```
