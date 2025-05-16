# Fetchfy - MCP Gateway Kubernetes Operator

Fetchfy is a Kubernetes Operator built in Go that orchestrates the integration of client/server services using the Model Context Protocol (MCP). It facilitates dynamic discovery, registration, and exposure of MCP tools and agents within a Kubernetes cluster.

## Description

The Fetchfy operator leverages Kubernetes' Gateway API to manage MCP (Model Context Protocol) services. It automatically detects services with specific labels and annotations, registers them dynamically in a central registry, and exposes them via a unified MCP Gateway endpoint.

### Features

- **Dynamic Service Discovery**: Automatically discovers MCP-enabled services in the cluster
- **Service Registry**: Maintains a registry of available MCP tools and agents
- **Gateway Management**: Creates and manages Gateway resources for MCP traffic
- **Metrics and Observability**: Includes Prometheus metrics for monitoring
- **High Availability**: Supports multiple gateways per cluster with leader election
- **Security**: Supports TLS for secure MCP communications

## Getting Started

### Prerequisites

- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/fetchfy:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/fetchfy:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
> privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/fetchfy:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/fetchfy/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
   can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Using MCP Integration

### Gateway CRD

The operator introduces a Custom Resource Definition (CRD) called `Gateway` that defines MCP gateways:

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
  enableTls: false
  tlsSecretRef: ""
```

### Integrating MCP Services

To integrate an MCP service with Fetchfy, apply the following labels and annotations to your Kubernetes service:

1. Apply the `mcp-enabled: "true"` label to your service
2. Use annotations to specify additional MCP information:
   - `mcp.fetchfy.ai/type`: either "tool" or "agent"
   - `mcp.fetchfy.ai/endpoint`: optional custom endpoint path

Example Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-mcp-tool
  labels:
    mcp-enabled: "true"
  annotations:
    mcp.fetchfy.ai/type: "tool"
    mcp.fetchfy.ai/endpoint: "/mcp/tools/my-tool"
spec:
  ports:
    - port: 8080
      targetPort: 8080
  selector:
    app: my-mcp-tool
```

The Fetchfy operator will automatically:

1. Discover this service based on the label
2. Register it in the MCP registry
3. Make it available through the MCP Gateway

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
