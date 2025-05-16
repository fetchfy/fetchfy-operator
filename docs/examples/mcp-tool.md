---
title: MCP Tool Integration Example
description: Step-by-step guide to integrate an MCP Tool with the Fetchfy Gateway Operator
---

# MCP Tool Integration Example

This guide demonstrates how to create and deploy an MCP Tool service that will be automatically discovered and registered by the Fetchfy Gateway Operator.

## What is an MCP Tool?

In the Model Context Protocol (MCP), a Tool is a specialized service that provides specific functionality to AI models. Tools have well-defined inputs and outputs and are designed to perform specific tasks such as:

- Data retrieval and transformation
- External API integration
- Numerical calculations
- Document processing
- Image generation or analysis

## Prerequisites

Before starting:

- Ensure the Fetchfy Operator is installed in your cluster
- Have a Gateway resource deployed (see [Basic Gateway Example](./basic-gateway.md))
- Familiarity with Kubernetes services and deployments

## Step 1: Create an MCP Tool Service

Let's create a simple calculator tool that can perform basic mathematical operations. We'll use Python with FastAPI to implement a compliant MCP tool.

### Tool Implementation

First, let's create the Python code for our calculator tool:

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, Any, List, Optional

app = FastAPI(title="MCP Calculator Tool")

class OperationRequest(BaseModel):
    parameters: Dict[str, Any]

class OperationResponse(BaseModel):
    result: Any

class ToolDefinition(BaseModel):
    name: str
    description: str
    parameters: Dict[str, Any]

@app.get("/mcp/definition")
async def get_definition() -> Dict[str, Any]:
    """Return the MCP tool definition"""
    return {
        "schema_version": "v1",
        "name": "calculator",
        "description": "A simple calculator tool for arithmetic operations",
        "tools": [
            {
                "name": "add",
                "description": "Add two numbers",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number"
                        }
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "subtract",
                "description": "Subtract second number from first",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number"
                        }
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "multiply",
                "description": "Multiply two numbers",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number"
                        }
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "divide",
                "description": "Divide first number by second",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number (cannot be zero)"
                        }
                    },
                    "required": ["a", "b"]
                }
            }
        ]
    }

@app.post("/mcp/tools/add")
async def add_numbers(request: OperationRequest) -> OperationResponse:
    """Add two numbers"""
    a = request.parameters.get("a")
    b = request.parameters.get("b")

    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise HTTPException(status_code=400, detail="Parameters 'a' and 'b' must be numbers")

    return OperationResponse(result=a + b)

@app.post("/mcp/tools/subtract")
async def subtract_numbers(request: OperationRequest) -> OperationResponse:
    """Subtract second number from first"""
    a = request.parameters.get("a")
    b = request.parameters.get("b")

    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise HTTPException(status_code=400, detail="Parameters 'a' and 'b' must be numbers")

    return OperationResponse(result=a - b)

@app.post("/mcp/tools/multiply")
async def multiply_numbers(request: OperationRequest) -> OperationResponse:
    """Multiply two numbers"""
    a = request.parameters.get("a")
    b = request.parameters.get("b")

    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise HTTPException(status_code=400, detail="Parameters 'a' and 'b' must be numbers")

    return OperationResponse(result=a * b)

@app.post("/mcp/tools/divide")
async def divide_numbers(request: OperationRequest) -> OperationResponse:
    """Divide first number by second"""
    a = request.parameters.get("a")
    b = request.parameters.get("b")

    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise HTTPException(status_code=400, detail="Parameters 'a' and 'b' must be numbers")

    if b == 0:
        raise HTTPException(status_code=400, detail="Cannot divide by zero")

    return OperationResponse(result=a / b)

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

### Dockerfile

Create a Dockerfile to containerize the tool:

```dockerfile
FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .

EXPOSE 8000

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

Create a requirements.txt file:

```
fastapi==0.104.1
uvicorn==0.24.0
pydantic==2.4.2
```

## Step 2: Build and Push the Container Image

Build and push the Docker image to a registry:

```bash
# Build the image
docker build -t your-registry/mcp-calculator:latest .

# Push the image
docker push your-registry/mcp-calculator:latest
```

## Step 3: Create Kubernetes Manifests

Create a deployment and service for the calculator tool:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calculator-tool
  namespace: mcp-example
  labels:
    app: calculator-tool
spec:
  replicas: 1
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
          image: your-registry/mcp-calculator:latest
          ports:
            - containerPort: 8000
          resources:
            limits:
              memory: "128Mi"
              cpu: "100m"
            requests:
              memory: "64Mi"
              cpu: "50m"
          readinessProbe:
            httpGet:
              path: /health
              port: 8000
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: calculator-tool
  namespace: mcp-example
  labels:
    app: calculator-tool
    mcp.fetchfy.io/type: tool # Important: This label enables discovery
  annotations:
    mcp.fetchfy.io/path: "/mcp/definition" # Optional: Path to MCP definition if not default
spec:
  selector:
    app: calculator-tool
  ports:
    - port: 80
      targetPort: 8000
  type: ClusterIP
```

Save this as `calculator-tool.yaml` and apply it:

```bash
kubectl apply -f calculator-tool.yaml
```

## Step 4: Verify Tool Registration

Check that the Gateway has registered the new tool:

```bash
kubectl get gateway basic-gateway -n mcp-example -o jsonpath='{.status.registeredServices}'
```

You should see the calculator tool in the list of registered services.

You can also describe the gateway to see more details:

```bash
kubectl describe gateway basic-gateway -n mcp-example
```

## Step 5: Test the Tool via the Gateway

To test the tool, you can port-forward the Gateway service:

```bash
kubectl port-forward -n mcp-example svc/basic-gateway 8080:8080
```

Now you can query the available tools:

```bash
curl http://localhost:8080/v1/tools
```

This should return the calculator tool and its operations.

To test a specific operation, for example, addition:

```bash
curl -X POST http://localhost:8080/v1/tools/calculator/add \
  -H "Content-Type: application/json" \
  -d '{"parameters": {"a": 5, "b": 3}}'
```

The response should be:

```json
{
  "result": 8
}
```

## Advanced Configuration

### Setting Tool Metadata

You can add additional metadata to your tool by using annotations:

```yaml
metadata:
  annotations:
    mcp.fetchfy.io/version: "1.0.0"
    mcp.fetchfy.io/description: "Simple calculator for basic arithmetic operations"
    mcp.fetchfy.io/maintainer: "example@example.com"
```

### Health Checks and Readiness

Ensure your tool has health check endpoints and implement Kubernetes probes:

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8000
  initialDelaySeconds: 5
  periodSeconds: 10
livenessProbe:
  httpGet:
    path: /health
    port: 8000
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Tool Authentication

If your tool requires authentication, you can set it up using Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: calculator-tool-auth
  namespace: mcp-example
type: Opaque
data:
  api-key: base64-encoded-api-key
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calculator-tool
spec:
  # ...
  template:
    spec:
      containers:
        - name: calculator
          # ...
          env:
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: calculator-tool-auth
                  key: api-key
```

## Common Issues and Troubleshooting

### Tool Not Registered

If your tool doesn't show up in the gateway:

1. Check the service labels:

```bash
kubectl get svc calculator-tool -n mcp-example -o yaml
```

Ensure it has the `mcp.fetchfy.io/type: tool` label.

2. Verify the tool is responding correctly:

```bash
kubectl port-forward -n mcp-example svc/calculator-tool 8000:80
curl http://localhost:8000/mcp/definition
```

3. Check Gateway operator logs:

```bash
kubectl logs -n fetchfy-system -l app=fetchfy-controller-manager
```

### Tool Registration but Operations Not Working

If the tool is registered but operations don't work:

1. Check the tool's log for errors:

```bash
kubectl logs -n mcp-example -l app=calculator-tool
```

2. Make direct requests to the tool to bypass the gateway:

```bash
kubectl port-forward -n mcp-example svc/calculator-tool 8000:80
curl -X POST http://localhost:8000/mcp/tools/add -H "Content-Type: application/json" -d '{"parameters": {"a": 5, "b": 3}}'
```

## Next Steps

- Explore [MCP Agent Integration](./mcp-agent.md) for more complex AI interactions
- Learn about implementing [Security](../guides/security.md) for your tools
- Set up [Monitoring](../guides/monitoring.md) for your MCP services
