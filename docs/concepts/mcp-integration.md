# MCP Integration

This page explains how the Model Context Protocol (MCP) is integrated within the Fetchfy operator and how to work with MCP services.

## What is the Model Context Protocol?

The Model Context Protocol (MCP) is a standardized communication protocol designed for AI systems, particularly large language models (LLMs) and intelligent agents, to interact with external tools and services. MCP allows:

- Tools to expose capabilities to AI systems
- Agents to perform actions in external systems
- Standardized communication between AI components

## MCP Service Types

In the Fetchfy ecosystem, MCP services come in two main types:

### 1. Tools

MCP Tools provide specific functionalities that can be invoked by clients. These are typically stateless services that perform a specific function and return results.

Examples include:

- Data retrieval tools
- Calculation services
- File manipulation tools
- External API integrations

### 2. Agents

MCP Agents are more complex services that can perform multi-step tasks, maintain state, and make autonomous decisions. They are typically used for complex workflows and delegated tasks.

Examples include:

- Task planning agents
- Persistent assistants
- Autonomous workflows
- Multi-step processing pipelines

## Registering MCP Services

To register a service with the Fetchfy MCP Gateway, you need to:

1. Add the `mcp-enabled: "true"` label to your Kubernetes Service
2. Use annotations to provide additional metadata

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-mcp-tool
  labels:
    mcp-enabled: "true"
  annotations:
    mcp.fetchfy.ai/type: "tool" # "tool" or "agent"
    mcp.fetchfy.ai/endpoint: "/mcp/tools/my-tool" # Optional: Custom endpoint
spec:
  # Service spec...
```

## MCP Service Requirements

For a service to be compatible with the Fetchfy MCP Gateway, it must:

1. Implement the MCP protocol (request/response format)
2. Listen on the port specified in the Kubernetes Service
3. Handle MCP function invocation requests
4. Return results in the MCP response format

## MCP Protocol Overview

The MCP protocol is HTTP-based and uses JSON for data exchange. The basic flow is:

1. **Discovery**: Clients request available tools/agents
2. **Invocation**: Clients send requests to invoke tools or delegate to agents
3. **Response**: Services return results or status updates

### Example MCP Request

```http
POST /mcp/tools/my-tool HTTP/1.1
Content-Type: application/json

{
  "function": "calculateDistance",
  "parameters": {
    "pointA": {"lat": 40.7128, "lng": -74.0060},
    "pointB": {"lat": 34.0522, "lng": -118.2437}
  }
}
```

### Example MCP Response

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "success",
  "result": {
    "distance": 3935.94,
    "unit": "km"
  }
}
```

## Implementing MCP Services

### Tool Implementation

Here's a simple example of a Go-based MCP tool implementation:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/mcp/tools/calculator", handleCalculatorRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCalculatorRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Function   string          `json:"function"`
		Parameters json.RawMessage `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result interface{}
	var err error

	switch req.Function {
	case "add":
		result, err = handleAdd(req.Parameters)
	case "subtract":
		result, err = handleSubtract(req.Parameters)
	default:
		http.Error(w, "Unknown function", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// Implementation of specific functions...
```

### Agent Implementation

Agents follow a similar pattern but typically maintain state and can handle multi-step interactions:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Session struct {
	ID    string
	State map[string]interface{}
}

var (
	sessions = map[string]*Session{}
	mu       sync.RWMutex
)

func main() {
	http.HandleFunc("/mcp/agents/assistant", handleAgentRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleAgentRequest(w http.ResponseWriter, r *http.Request) {
	// Handle session creation, continuation, or termination
	// Process agent interactions
	// Update session state
	// Return appropriate responses
}

// Agent-specific logic...
```

## Using MCP Clients

To connect to the Fetchfy MCP Gateway from your applications, you can use MCP client libraries:

### Go Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	mcp "github.com/modelcontextprotocol/client-go"
)

func main() {
	client, err := mcp.NewClient("http://fetchfy-gateway:8080")
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}

	// List available tools
	tools, err := client.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Println("Available tools:")
	for _, tool := range tools {
		fmt.Printf("- %s\n", tool.Name)
	}

	// Invoke a tool
	result, err := client.InvokeTool(context.Background(), "calculator", "add", map[string]interface{}{
		"a": 5,
		"b": 3,
	})

	if err != nil {
		log.Fatalf("Failed to invoke tool: %v", err)
	}

	fmt.Printf("Result: %v\n", result)
}
```

## Best Practices

When implementing MCP services for use with Fetchfy:

1. **Versioning**: Include version information in your MCP responses
2. **Documentation**: Provide clear documentation of available functions
3. **Error Handling**: Return detailed error messages with appropriate HTTP status codes
4. **Rate Limiting**: Implement rate limiting for resource-intensive operations
5. **Monitoring**: Add logging and metrics for observability
6. **Validation**: Validate input parameters thoroughly
7. **Idempotency**: Make operations idempotent where appropriate

## Advanced Topics

### Authentication & Authorization

For secure environments, you can implement authentication for MCP services:

1. **API Keys**: Include API key validation in your MCP services
2. **JWT Tokens**: Use JWT for authentication and authorization
3. **Service Accounts**: Leverage Kubernetes service accounts for authentication

### Streaming Responses

Some MCP operations benefit from streaming responses, especially for long-running operations:

```go
func handleStreamingRequest(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Stream data as it becomes available
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "data: {\"progress\": %d}\n\n", i*10)
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Fprintf(w, "data: {\"status\": \"complete\"}\n\n")
	flusher.Flush()
}
```
