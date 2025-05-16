# Quick Start

This guide will help you set up a complete working example of the Fetchfy MCP Gateway Operator in just a few minutes.

## Prerequisites

- Kubernetes cluster (v1.19+)
- `kubectl` installed and configured to communicate with your cluster
- Cluster admin privileges (for CRD installation)

## Step 1: Install Fetchfy Operator

First, install the Fetchfy operator in your cluster:

```bash
kubectl apply -f https://github.com/fetchfy/fetchfy-operator/releases/latest/download/fetchfy-operator.yaml
```

Wait for the operator to be ready:

```bash
kubectl -n fetchfy-system wait --for=condition=Available deployment/fetchfy-controller-manager --timeout=60s
```

## Step 2: Create a Gateway

Create a file named `gateway.yaml` with the following content:

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: fetchfy-gateway
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
```

Apply the Gateway resource:

```bash
kubectl apply -f gateway.yaml
```

## Step 3: Create an MCP Service

Let's create a sample MCP tool service. Create a file named `mcp-tool.yaml` with the following content:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mcp-tool-example
  labels:
    app: mcp-tool-example
    mcp-enabled: "true"
  annotations:
    mcp.fetchfy.ai/type: "tool"
    mcp.fetchfy.ai/endpoint: "/mcp/tools/example"
spec:
  ports:
    - port: 8080
      targetPort: 8080
      protocol: TCP
      name: http
  selector:
    app: mcp-tool-example
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-tool-example
  labels:
    app: mcp-tool-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mcp-tool-example
  template:
    metadata:
      labels:
        app: mcp-tool-example
    spec:
      containers:
        - name: mcp-tool
          image: example/mcp-tool:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              name: http
```

Apply the MCP tool service:

```bash
kubectl apply -f mcp-tool.yaml
```

## Step 4: Verify Registration

Check that the MCP service has been registered with the Gateway:

```bash
kubectl get gateway fetchfy-gateway -o jsonpath="{.status.mcpServices}"
```

You should see output similar to the following:

```json
[
  {
    "name": "mcp-tool-example",
    "namespace": "default",
    "type": "tool",
    "endpoint": "/mcp/tools/example",
    "status": "Available",
    "lastUpdated": "2025-05-16T10:30:45Z"
  }
]
```

## Step 5: Connect to the MCP Gateway

To use the MCP Gateway in your applications, you'll need to connect to it. Here's a simple example in Go:

```go
package main

import (
    "context"
    "fmt"
    "log"

    mcp "github.com/modelcontextprotocol/client-go"
)

func main() {
    client, err := mcp.NewClient("http://fetchfy-gateway.default.svc.cluster.local:8080")
    if err != nil {
        log.Fatalf("Failed to create MCP client: %v", err)
    }

    // List available tools
    tools, err := client.ListTools(context.Background())
    if err != nil {
        log.Fatalf("Failed to list tools: %v", err)
    }

    fmt.Println("Available MCP tools:")
    for _, tool := range tools {
        fmt.Printf("- %s\n", tool.Name)
    }
}
```

## What's Next?

Now that you have a working Fetchfy setup, you can:

- [Create more complex Gateway configurations](../guides/creating-gateway.md)
- [Deploy different types of MCP services](../guides/deploying-services.md)
- [Set up monitoring](../guides/monitoring.md)
- [Enable TLS for secure communication](../guides/security.md)
