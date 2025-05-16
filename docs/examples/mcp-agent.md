---
title: MCP Agent Integration Example
description: Step-by-step guide to integrate an MCP Agent with the Fetchfy Gateway Operator
---

# MCP Agent Integration Example

This guide demonstrates how to create and deploy an MCP Agent service that will be automatically discovered and registered by the Fetchfy Gateway Operator.

## What is an MCP Agent?

In the Model Context Protocol (MCP), an Agent is a more complex service compared to a Tool. While Tools provide specific functionality with well-defined inputs and outputs, Agents can:

- Maintain state across interactions
- Perform multi-step reasoning
- Execute complex workflows
- Coordinate multiple tools
- Handle conversation history and context

Agents typically implement the full MCP protocol with support for both synchronous and asynchronous operations.

## Prerequisites

Before starting:

- Ensure the Fetchfy Operator is installed in your cluster
- Have a Gateway resource deployed (see [Basic Gateway Example](./basic-gateway.md))
- Familiarity with Kubernetes services and deployments

## Step 1: Create an MCP Agent Service

Let's create a simple research agent that can perform web searches, summarize content, and maintain context across interactions. We'll use Python with FastAPI to implement a compliant MCP agent.

### Agent Implementation

First, let's create the Python code for our research agent:

```python
from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel
from typing import Dict, Any, List, Optional, Union
import uuid
import time
import json

app = FastAPI(title="MCP Research Agent")

# In-memory storage for agent state
sessions = {}

class AgentRequest(BaseModel):
    parameters: Dict[str, Any]
    context: Optional[Dict[str, Any]] = None

class AgentResponse(BaseModel):
    result: Any
    context: Optional[Dict[str, Any]] = None

class AsyncTaskResponse(BaseModel):
    task_id: str
    status: str = "pending"

class AgentDefinition(BaseModel):
    schema_version: str
    name: str
    description: str
    agent_type: str
    capabilities: List[Dict[str, Any]]

# Mock data for demonstration
SEARCH_RESULTS = {
    "ai": ["Machine learning breakthrough announced", "New AI regulation proposed", "AI adoption reaches record high"],
    "climate": ["Latest climate research findings", "New carbon capture technology", "Climate policy update"],
    "technology": ["Tech sector growth continues", "New smartphone features unveiled", "Quantum computing milestone"]
}

@app.get("/mcp/definition")
async def get_definition() -> Dict[str, Any]:
    """Return the MCP agent definition"""
    return {
        "schema_version": "v1",
        "name": "research_agent",
        "description": "Research agent for web search and content analysis",
        "agent_type": "assistant",
        "capabilities": [
            {
                "name": "search",
                "description": "Search for information on a topic",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "Search query"
                        },
                        "max_results": {
                            "type": "integer",
                            "description": "Maximum number of results to return",
                            "default": 3
                        }
                    },
                    "required": ["query"]
                }
            },
            {
                "name": "summarize",
                "description": "Summarize previous search results",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "style": {
                            "type": "string",
                            "description": "Summarization style (brief, detailed)",
                            "enum": ["brief", "detailed"],
                            "default": "brief"
                        }
                    }
                }
            },
            {
                "name": "analyze",
                "description": "Run a complex analysis on a topic (async operation)",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "topic": {
                            "type": "string",
                            "description": "Topic to analyze"
                        },
                        "depth": {
                            "type": "string",
                            "description": "Analysis depth",
                            "enum": ["shallow", "medium", "deep"],
                            "default": "medium"
                        }
                    },
                    "required": ["topic"]
                },
                "async": true
            }
        ]
    }

@app.post("/mcp/agent/search")
async def search(request: AgentRequest) -> AgentResponse:
    """Search for information on a topic"""
    query = request.parameters.get("query", "").lower()
    max_results = min(request.parameters.get("max_results", 3), 10)  # Limit to 10 max

    if not query:
        raise HTTPException(status_code=400, detail="Query parameter is required")

    # Get session context or create new
    context = request.context or {}
    session_id = context.get("session_id", str(uuid.uuid4()))

    # Find matching results
    results = []
    for key, values in SEARCH_RESULTS.items():
        if key in query or query in key:
            results.extend(values)

    # Limit results
    results = results[:max_results] if results else ["No results found for: " + query]

    # Update session state
    if session_id not in sessions:
        sessions[session_id] = {"searches": []}

    sessions[session_id]["searches"].append({
        "query": query,
        "results": results,
        "timestamp": time.time()
    })

    # Return results and updated context
    return AgentResponse(
        result=results,
        context={
            "session_id": session_id,
            "last_query": query,
            "result_count": len(results)
        }
    )

@app.post("/mcp/agent/summarize")
async def summarize(request: AgentRequest) -> AgentResponse:
    """Summarize previous search results"""
    context = request.context or {}
    session_id = context.get("session_id")
    style = request.parameters.get("style", "brief")

    if not session_id or session_id not in sessions:
        raise HTTPException(status_code=400, detail="No active session found. Please perform a search first.")

    session = sessions[session_id]
    if not session.get("searches"):
        raise HTTPException(status_code=400, detail="No searches found in this session")

    # Create summary based on previous searches
    searches = session["searches"]
    if style == "brief":
        summary = f"Based on {len(searches)} searches, key topics include: " + \
                 ", ".join(search["query"] for search in searches[-3:])
    else:
        summary = "Detailed summary of research:\n\n"
        for i, search in enumerate(searches):
            summary += f"Search {i+1}: '{search['query']}'\n"
            summary += "Results:\n"
            for j, result in enumerate(search["results"]):
                summary += f"- {result}\n"
            summary += "\n"

    return AgentResponse(
        result=summary,
        context={
            "session_id": session_id,
            "summary_style": style,
            "search_count": len(searches)
        }
    )

def run_analysis(topic, depth, task_id):
    """Background task for analysis"""
    # Simulate a long-running process
    time.sleep(10)

    # Create analysis results based on depth
    if depth == "shallow":
        analysis = f"Basic analysis of {topic}: General trends identified."
    elif depth == "deep":
        analysis = f"Deep analysis of {topic}: Comprehensive insights with statistical breakdowns and future projections."
    else:  # medium
        analysis = f"Medium-depth analysis of {topic}: Key findings and moderate insights provided."

    # Store the result
    tasks[task_id] = {
        "status": "completed",
        "result": analysis,
        "completed_at": time.time()
    }

# In-memory task storage
tasks = {}

@app.post("/mcp/agent/analyze")
async def analyze(request: AgentRequest, background_tasks: BackgroundTasks) -> AsyncTaskResponse:
    """Run a complex analysis on a topic (async operation)"""
    topic = request.parameters.get("topic")
    depth = request.parameters.get("depth", "medium")

    if not topic:
        raise HTTPException(status_code=400, detail="Topic parameter is required")

    # Create a task ID
    task_id = str(uuid.uuid4())

    # Store initial task state
    tasks[task_id] = {
        "status": "pending",
        "created_at": time.time(),
        "topic": topic,
        "depth": depth
    }

    # Run analysis in background
    background_tasks.add_task(run_analysis, topic, depth, task_id)

    return AsyncTaskResponse(task_id=task_id)

@app.get("/mcp/tasks/{task_id}")
async def get_task_status(task_id: str) -> Dict[str, Any]:
    """Check status of an async task"""
    if task_id not in tasks:
        raise HTTPException(status_code=404, detail="Task not found")

    task = tasks[task_id]
    response = {
        "task_id": task_id,
        "status": task["status"]
    }

    if task["status"] == "completed":
        response["result"] = task["result"]

    return response

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

### Dockerfile

Create a Dockerfile to containerize the agent:

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
docker build -t your-registry/mcp-research-agent:latest .

# Push the image
docker push your-registry/mcp-research-agent:latest
```

## Step 3: Create Kubernetes Manifests

Create a deployment and service for the research agent:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: research-agent
  namespace: mcp-example
  labels:
    app: research-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: research-agent
  template:
    metadata:
      labels:
        app: research-agent
    spec:
      containers:
        - name: agent
          image: your-registry/mcp-research-agent:latest
          ports:
            - containerPort: 8000
          resources:
            limits:
              memory: "256Mi"
              cpu: "200m"
            requests:
              memory: "128Mi"
              cpu: "100m"
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
---
apiVersion: v1
kind: Service
metadata:
  name: research-agent
  namespace: mcp-example
  labels:
    app: research-agent
    mcp.fetchfy.io/type: agent # Important: This label identifies it as an agent
  annotations:
    mcp.fetchfy.io/path: "/mcp/definition" # Optional: Path to MCP definition if not default
spec:
  selector:
    app: research-agent
  ports:
    - port: 80
      targetPort: 8000
  type: ClusterIP
```

Save this as `research-agent.yaml` and apply it:

```bash
kubectl apply -f research-agent.yaml
```

## Step 4: Update the Gateway Configuration

Ensure your Gateway is configured to discover agents by updating the service selector:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: mcp-gateway
  namespace: mcp-example
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: "agent" # Add agent type
  # ...other settings
```

Alternatively, you can use a more inclusive selector:

```yaml
serviceSelector:
  matchExpressions:
    - key: mcp.fetchfy.io/type
      operator: In
      values: ["tool", "agent"]
```

Apply the updated Gateway configuration:

```bash
kubectl apply -f gateway.yaml
```

## Step 5: Verify Agent Registration

Check that the Gateway has registered the new agent:

```bash
kubectl get gateway mcp-gateway -n mcp-example -o jsonpath='{.status.registeredServices}'
```

You should see the research agent in the list of registered services.

Describe the gateway to see more details:

```bash
kubectl describe gateway mcp-gateway -n mcp-example
```

## Step 6: Test the Agent via the Gateway

To test the agent, port-forward the Gateway service:

```bash
kubectl port-forward -n mcp-example svc/mcp-gateway 8080:8080
```

Query the available agents:

```bash
curl http://localhost:8080/v1/agents
```

This should return the research agent and its capabilities.

### Testing Stateful Interactions

Agents maintain state across interactions. Let's test this:

1. First, perform a search:

```bash
curl -X POST http://localhost:8080/v1/agents/research_agent/search \
  -H "Content-Type: application/json" \
  -d '{"parameters": {"query": "ai", "max_results": 2}}'
```

Response (note the session ID in the context):

```json
{
  "result": [
    "Machine learning breakthrough announced",
    "New AI regulation proposed"
  ],
  "context": {
    "session_id": "f8e7d6c5-b4a3-42f1-9e8d-7c6b5a4f3d2e",
    "last_query": "ai",
    "result_count": 2
  }
}
```

2. Then, use the returned context to get a summary:

```bash
curl -X POST http://localhost:8080/v1/agents/research_agent/summarize \
  -H "Content-Type: application/json" \
  -d '{"parameters": {"style": "brief"}, "context": {"session_id": "f8e7d6c5-b4a3-42f1-9e8d-7c6b5a4f3d2e"}}'
```

Response:

```json
{
  "result": "Based on 1 searches, key topics include: ai",
  "context": {
    "session_id": "f8e7d6c5-b4a3-42f1-9e8d-7c6b5a4f3d2e",
    "summary_style": "brief",
    "search_count": 1
  }
}
```

### Testing Asynchronous Operations

Let's test an asynchronous operation:

1. Start the analysis:

```bash
curl -X POST http://localhost:8080/v1/agents/research_agent/analyze \
  -H "Content-Type: application/json" \
  -d '{"parameters": {"topic": "climate change", "depth": "deep"}}'
```

Response:

```json
{
  "task_id": "a1b2c3d4-e5f6-7a8b-9c0d-123456789abc",
  "status": "pending"
}
```

2. Check the task status using the task ID:

```bash
curl http://localhost:8080/v1/tasks/a1b2c3d4-e5f6-7a8b-9c0d-123456789abc
```

Initial response:

```json
{
  "task_id": "a1b2c3d4-e5f6-7a8b-9c0d-123456789abc",
  "status": "pending"
}
```

3. Check again after a few seconds:

```bash
curl http://localhost:8080/v1/tasks/a1b2c3d4-e5f6-7a8b-9c0d-123456789abc
```

Final response:

```json
{
  "task_id": "a1b2c3d4-e5f6-7a8b-9c0d-123456789abc",
  "status": "completed",
  "result": "Deep analysis of climate change: Comprehensive insights with statistical breakdowns and future projections."
}
```

## Key Differences Between Tools and Agents

| Feature           | MCP Tool        | MCP Agent           |
| ----------------- | --------------- | ------------------- |
| Stateful          | No              | Yes                 |
| Maintains context | No              | Yes                 |
| Async operations  | Rarely          | Common              |
| Complexity        | Simple, focused | Complex, multi-step |
| Use case          | Single tasks    | Complex workflows   |

## Advanced Agent Configuration

### Scaling for Production

For production use, consider scaling and persistence:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: research-agent
  namespace: mcp-example
spec:
  replicas: 3
  serviceName: research-agent
  selector:
    matchLabels:
      app: research-agent
  template:
    # ...container spec
  volumeClaimTemplates:
    - metadata:
        name: agent-data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
```

### Persistent Storage for Agent State

For real agents, you'll want persistent storage for the session data:

```python
# Use Redis, MongoDB, or another database instead of in-memory storage
import redis

redis_client = redis.Redis(host="redis", port=6379)

def get_session(session_id):
    data = redis_client.get(f"session:{session_id}")
    return json.loads(data) if data else None

def save_session(session_id, data):
    redis_client.set(f"session:{session_id}", json.dumps(data))
```

### Authentication and Authorization

Add authentication for your agent:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: agent-credentials
  namespace: mcp-example
type: Opaque
data:
  api-key: base64-encoded-api-key
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: research-agent
spec:
  # ...
  template:
    spec:
      containers:
        - name: agent
          # ...
          env:
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: agent-credentials
                  key: api-key
```

## Common Issues and Troubleshooting

### Agent Not Registered

If your agent isn't showing up in the gateway:

1. Check the service labels:

```bash
kubectl get svc research-agent -n mcp-example -o yaml
```

Ensure it has the `mcp.fetchfy.io/type: agent` label.

2. Verify the agent responds correctly:

```bash
kubectl port-forward -n mcp-example svc/research-agent 8000:80
curl http://localhost:8000/mcp/definition
```

3. Check Gateway operator logs:

```bash
kubectl logs -n fetchfy-system -l app=fetchfy-controller-manager
```

### Context Not Maintained Between Calls

If the agent's state isn't preserved between calls:

1. Check if the context is being properly passed in requests
2. Verify the agent's session storage is working
3. Inspect the agent logs for any session-related errors

```bash
kubectl logs -n mcp-example -l app=research-agent
```

### Async Tasks Not Completing

If asynchronous tasks don't complete:

1. Check if the background task processing is functioning
2. Look for errors in the agent logs
3. Verify the agent has sufficient resources to process tasks

## Next Steps

- Implement a more sophisticated agent with real external APIs
- Connect your agent to other MCP tools via the gateway
- Set up [Monitoring](../guides/monitoring.md) for your agent
- Explore [Security](../guides/security.md) options for protecting agent communications
