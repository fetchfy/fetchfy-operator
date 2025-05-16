# Fetchfy MCP Gateway Operator

![Fetchfy Logo](images/logo.svg){ align=right width=150 }

Fetchfy is a Kubernetes Operator built in Go that orchestrates the integration of client/server services using the Model Context Protocol (MCP). It facilitates dynamic discovery, registration, and exposure of MCP tools and agents within a Kubernetes cluster.

## What is MCP?

The Model Context Protocol (MCP) is a protocol for communication between intelligent agents and external tools or services. Fetchfy bridges the gap between the MCP ecosystem and Kubernetes, making it easy to deploy, discover, and use MCP-enabled services.

## Key Features

- **Dynamic Service Discovery**: Automatically discover MCP-enabled services in your cluster
- **Centralized Registry**: Maintain a central registry of available MCP tools and agents
- **Gateway Management**: Create and manage Gateway resources for MCP traffic
- **Observability**: Prometheus metrics for comprehensive monitoring
- **High Availability**: Support for multiple gateways per cluster with leader election
- **Security**: TLS support for secure MCP communications

## Why Fetchfy?

Fetchfy solves the problem of integrating and orchestrating MCP services in a Kubernetes environment. It eliminates the need for manual configuration and service discovery, allowing you to focus on building intelligent applications.

With Fetchfy, you can:

- **Simplify Deployment**: Easily deploy and manage MCP services
- **Automate Discovery**: Automatically discover and register MCP tools and agents
- **Enhance Observability**: Monitor MCP traffic and service health
- **Secure Communication**: Enable TLS for secure MCP communications

## Quick Start

Get up and running quickly with our [Quick Start Guide](getting-started/quickstart.md).

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

## Compatibility

Fetchfy is compatible with:

- Kubernetes 1.19+
- MCP Protocol v1.0+

## Community & Support

- [GitHub Issues](https://github.com/fetchfy/fetchfy-operator/issues)
- [Contribution Guidelines](development/contributing.md)
