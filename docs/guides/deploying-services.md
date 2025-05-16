# Deploying MCP Services

This guide explains how to deploy MCP-enabled services that can be automatically discovered and registered with the Fetchfy Gateway.

## Prerequisites

- Kubernetes cluster with Fetchfy operator installed
- At least one Gateway resource deployed
- Basic understanding of MCP (Model Context Protocol)

## MCP Service Types

Fetchfy supports two types of MCP services:

1. **Tools**: Stateless services that provide specific functionalities
2. **Agents**: Stateful services that can perform complex, multi-step tasks

## General Deployment Process

The deployment process consists of these core steps:

1. Create a Docker image for your MCP service
2. Deploy the service to Kubernetes
3. Label it as MCP-enabled
4. Add appropriate annotations for service configuration

## Service Requirements

For a service to be properly registered and used with Fetchfy, it must:

1. Implement the MCP protocol
2. Listen on the correct port
3. Handle HTTP/JSON requests according to the MCP specification
4. Return responses in the expected format

## Deploying an MCP Tool

### 1. Create Deployment

Create a YAML file (e.g., `mcp-tool.yaml`) for your tool service:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calculator-tool
  labels:
    app: calculator-tool
spec:
  replicas: 2
  selector:
    matchLabels:
      app: calculator-tool
  template:
    metadata:
      labels:
        app: calculator-tool
    spec:
      containers:
        - name: calculator
          image: myregistry/calculator-mcp:v1.0.0
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
```

### 2. Create Service with MCP Labels

Create a Kubernetes Service with the required MCP labels and annotations:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: calculator-tool
  labels:
    app: calculator-tool
    mcp-enabled: "true" # This label is required for Fetchfy discovery
  annotations:
    mcp.fetchfy.ai/type: "tool" # Specifies this is a tool service
    mcp.fetchfy.ai/endpoint: "/mcp/tools/calculator" # Optional: Custom endpoint path
spec:
  selector:
    app: calculator-tool
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
```

### 3. Apply the Configuration

Apply both resources to your cluster:

```bash
kubectl apply -f mcp-tool.yaml
```

### 4. Verify Registration

Check that the service has been registered with the Gateway:

```bash
kubectl get gateway <gateway-name> -o jsonpath="{.status.mcpServices[*].name}"
```

You should see your service name in the output.

## Deploying an MCP Agent

The process for deploying an MCP Agent is similar, with a few key differences:

### 1. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: assistant-agent
  labels:
    app: assistant-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: assistant-agent
  template:
    metadata:
      labels:
        app: assistant-agent
    spec:
      containers:
        - name: assistant
          image: myregistry/assistant-mcp:v1.0.0
          ports:
            - containerPort: 8080
```

### 2. Create Service with MCP Labels

```yaml
apiVersion: v1
kind: Service
metadata:
  name: assistant-agent
  labels:
    app: assistant-agent
    mcp-enabled: "true"
  annotations:
    mcp.fetchfy.ai/type: "agent" # Specifies this is an agent service
    mcp.fetchfy.ai/endpoint: "/mcp/agents/assistant"
spec:
  selector:
    app: assistant-agent
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
```

### 3. Apply and Verify

Same as for tools.

## Configuration Options

### Available Annotations

| Annotation                   | Description                     | Default                   |
| ---------------------------- | ------------------------------- | ------------------------- |
| `mcp.fetchfy.ai/type`        | Service type: "tool" or "agent" | "tool"                    |
| `mcp.fetchfy.ai/endpoint`    | Custom endpoint path            | `/mcp/{namespace}/{name}` |
| `mcp.fetchfy.ai/description` | Human-readable description      | None                      |
| `mcp.fetchfy.ai/version`     | Version information             | None                      |
| `mcp.fetchfy.ai/timeout`     | Request timeout in seconds      | `60`                      |

## Best Practices

### Resource Management

Properly configure resource requests and limits:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

### Health Checks

Implement health and readiness probes:

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Horizontal Scaling

For services that need to handle high load, configure horizontal scaling:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: calculator-tool-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: calculator-tool
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

## Advanced Configuration

### Secure Communication

To enable secure communication, use HTTPS and configure the Gateway with TLS.

1. Create a TLS secret:

```bash
kubectl create secret tls mcp-tls-secret \
  --cert=path/to/cert.crt \
  --key=path/to/private.key
```

2. Configure the Gateway to use TLS:

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: secure-gateway
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
  enableTls: true
  tlsSecretRef: "mcp-tls-secret"
```

### Access Control

You can use Kubernetes Network Policies to control access to your MCP services:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-service-policy
spec:
  podSelector:
    matchLabels:
      app: calculator-tool
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: fetchfy-system
          podSelector:
            matchLabels:
              app.kubernetes.io/name: fetchfy-operator
      ports:
        - protocol: TCP
          port: 8080
```

## Troubleshooting

### Service Not Registering

If your service doesn't register with the Gateway:

1. **Check labels**: Ensure the service has the `mcp-enabled: "true"` label
2. **Check Gateway selector**: Verify that your service matches the Gateway's `serviceSelector`
3. **Check service status**: Look at the Gateway status for any error messages
4. **Check operator logs**: Examine logs for the Fetchfy operator

```bash
kubectl logs -n fetchfy-system deployment/fetchfy-operator-controller-manager
```

### Service Registered but Not Accessible

If your service is registered but not accessible:

1. **Check service health**: Ensure your service pods are running and ready
2. **Check network connectivity**: Verify network policies and service connectivity
3. **Check service implementation**: Ensure it correctly implements the MCP protocol
4. **Check logs**: Look for errors in both the operator and service logs
