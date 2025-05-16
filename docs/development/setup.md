---
title: Development Setup
description: Set up your development environment for the Fetchfy MCP Gateway Operator
---

# Development Setup

This guide will help you set up your development environment to work on the Fetchfy MCP Gateway Operator.

## Prerequisites

Before you begin, ensure you have the following installed:

- Go (version 1.20 or later)
- Docker
- kubectl
- A Kubernetes cluster (local or remote)
- make
- git

## Setting Up Your Environment

### Clone the Repository

```bash
git clone https://github.com/fetchfy/fetchfy-operator.git
cd fetchfy-operator
```

### Installing Dependencies

The project uses Go modules for dependency management. To install all dependencies:

```bash
go mod download
```

### Development Tools

The project relies on several development tools:

- [controller-gen](https://github.com/kubernetes-sigs/controller-tools) for generating CRD manifests
- [kustomize](https://github.com/kubernetes-sigs/kustomize) for customizing Kubernetes manifests
- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) for scaffolding and building operators

Install these tools using:

```bash
# controller-gen
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

# kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
sudo mv kustomize /usr/local/bin/

# kubebuilder
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

## Local Development Workflow

### Building the Operator

To build the operator locally:

```bash
make build
```

This will create the operator binary in the `bin/` directory.

### Running Tests

Run the unit tests:

```bash
make test
```

For integration tests (requires a Kubernetes cluster):

```bash
make test-integration
```

### Running the Operator Locally

You can run the operator outside of the cluster for easier debugging:

```bash
make install # Install CRDs in the cluster
make run # Run the controller locally
```

### Building and Pushing a Docker Image

To build and push a Docker image of the operator:

```bash
make docker-build docker-push IMG=your-registry/fetchfy-operator:tag
```

### Deploying to a Cluster

Deploy the operator to your Kubernetes cluster:

```bash
make deploy IMG=your-registry/fetchfy-operator:tag
```

## Development Environment

### Using DevContainer

This project includes a `.devcontainer` configuration for Visual Studio Code. To use it:

1. Install the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. Open the project in VS Code
3. Click the green button in the bottom-left corner and select "Reopen in Container"

### Setting Up Kind for Local Development

[Kind](https://kind.sigs.k8s.io/) is a tool for running local Kubernetes clusters using Docker:

```bash
# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.17.0/kind-$(uname)-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/

# Create a cluster
kind create cluster --name fetchfy-dev

# Set kubectl context
kubectl cluster-info --context kind-fetchfy-dev

# Deploy the operator
make deploy IMG=your-registry/fetchfy-operator:tag
```

### Using Skaffold for Development

[Skaffold](https://skaffold.dev/) can simplify the development workflow:

```bash
# Install Skaffold
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-$(uname -s)-$(uname -m)
chmod +x skaffold
sudo mv skaffold /usr/local/bin/

# Run development mode
skaffold dev --port-forward
```

## Project Structure

The project follows the standard Kubebuilder layout:

- `api/`: API definitions (CRDs)
- `cmd/`: Entry point for the operator
- `config/`: Kubernetes manifests
- `controllers/`: Controller implementations
- `pkg/`: Shared packages and utilities
- `internal/`: Internal implementation details
- `docs/`: Documentation
- `hack/`: Scripts and tools

## Code Generation

### Generating CRDs

After modifying the API types in `api/`, regenerate the CRDs:

```bash
make manifests
```

### Generating DeepCopy Methods

Generate deepcopy methods for API types:

```bash
make generate
```

## Troubleshooting

### Common Issues

- **Authentication Issues**: Ensure your Docker registry credentials are set up correctly
- **CRD Installation Failures**: Check if the CRDs already exist and delete them if necessary
- **Controller Permission Errors**: Verify the RBAC permissions in `config/rbac/`

### Debugging

- Use `kubectl logs` to check operator logs
- For local debugging, use Go's debugging tools (`dlv`)
- Add temporary debug logs with `ctrl.Log.Info()`

## Next Steps

- Read the [Contributing Guide](./contributing.md) for contribution guidelines
- Check out the [Debugging Guide](./debugging.md) for more advanced debugging techniques
