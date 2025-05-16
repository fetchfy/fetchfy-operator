---
title: Overview
description: Key concepts behind the Fetchfy MCP Gateway Operator
---

# Overview

The Fetchfy MCP Gateway Operator is designed to simplify the integration of AI model tools and agents in a Kubernetes environment by implementing the Model Context Protocol (MCP). This page provides an overview of the key concepts behind the operator.

## What is the Model Context Protocol (MCP)?

The Model Context Protocol (MCP) is a standardized communication protocol designed for interaction between AI models, tools, and agents. It enables:

- **Dynamic tool discovery**: Models can discover what tools are available
- **Parameter validation**: Tools can define their parameter requirements
- **Structured communication**: Standardized request/response formats
- **Runtime integration**: Tools and agents can be easily integrated at runtime

MCP allows AI models to interact with a wide range of tools in a consistent, discoverable manner, enabling them to perform complex tasks that require external functionality.

## Key Components of the Fetchfy Operator

The Fetchfy Operator introduces several key components to manage MCP integration in Kubernetes:

### Gateway Custom Resource

The Gateway Custom Resource Definition (CRD) is the primary interface for configuring MCP integration. It defines:

- How MCP services should be discovered in the cluster
- Which port the MCP Gateway should listen on
- Security settings like TLS configuration
- Connection settings and other parameters

### MCP Server

The MCP Server component implements the Model Context Protocol and exposes an API endpoint for models to discover and interact with tools and agents. It:

- Handles MCP protocol requests
- Routes requests to appropriate services
- Manages tool registration and deregistration
- Implements protocol validation

### Service Registry

The Service Registry maintains an up-to-date catalog of all MCP-compatible tools and agents in the cluster. It:

- Tracks service availability
- Manages metadata about each service
- Provides quick lookups for the MCP server
- Handles service lifecycle events

### Service Watcher

The Service Watcher monitors the Kubernetes API server for changes to services that match the configured selectors. It:

- Detects new MCP-compatible services
- Removes services that are no longer available
- Updates the registry when service details change
- Applies filtering based on configuration

## Service Types

The Fetchfy Operator works with two primary types of MCP services:

### Tools

MCP Tools represent specific capabilities that models can use. Examples include:

- Code generators
- Data extractors
- Image processors
- API integrations
- Mathematical functions

Tools expose a defined set of parameters and return structured responses.

### Agents

MCP Agents are more complex entities that can perform multi-step reasoning or maintain state across interactions. Examples include:

- Complex workflow engines
- Specialized assistants
- Domain-specific reasoners

## How It Works

1. The operator deploys an MCP Gateway based on the Gateway resource configuration
2. The Service Watcher identifies services with appropriate labels/annotations
3. Identified services are registered in the Service Registry
4. The MCP Server exposes these services via the MCP Protocol
5. AI models connect to the gateway and discover available tools and agents
6. Models can then invoke these tools and agents as needed

## Benefits of Using Fetchfy

- **Simplified integration**: Easy registration and discovery of MCP-compatible services
- **Kubernetes-native**: Leverages Kubernetes concepts like services and selectors
- **Scalable**: Designed for both small and large-scale deployments
- **Secure**: Built-in support for TLS and authentication
- **Observable**: Integrated metrics and monitoring
- **Declarative**: Configuration through Kubernetes resources

## Next Steps

- Learn about the detailed [Architecture](./architecture.md)
- Explore [MCP Integration](./mcp-integration.md) specifics
- Follow the [Installation Guide](../getting-started/installation.md) to deploy the operator
