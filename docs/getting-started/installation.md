# Installation

This guide walks you through the process of installing the Fetchfy MCP Gateway Operator in your Kubernetes cluster.

## Prerequisites

Before installing Fetchfy, ensure that you have:

- A Kubernetes cluster (v1.19+)
- `kubectl` installed and configured to communicate with your cluster
- Cluster admin privileges (for CRD installation)

## Installation Methods

There are several ways to install Fetchfy. Choose the method that best fits your environment.

### Method 1: Using Kubectl

This is the simplest way to deploy Fetchfy to your cluster:

```bash
kubectl apply -f https://github.com/fetchfy/fetchfy-operator/releases/latest/download/fetchfy-operator.yaml
```

This will install:

- Custom Resource Definitions (CRDs)
- RBAC policies
- The Fetchfy operator deployment

### Method 2: Using Helm

If you prefer using Helm, you can install Fetchfy as follows:

```bash
helm repo add fetchfy https://fetchfy.github.io/charts
helm repo update
helm install fetchfy-operator fetchfy/fetchfy-operator
```

### Method 3: From Source

For development or customization purposes, you can deploy from source:

1. Clone the repository:

   ```bash
   git clone https://github.com/fetchfy/fetchfy-operator.git
   cd fetchfy-operator
   ```

2. Install the CRDs:

   ```bash
   make install
   ```

3. Deploy the operator:
   ```bash
   make deploy
   ```

## Verifying the Installation

To verify that the Fetchfy operator is running correctly:

```bash
kubectl get pods -n fetchfy-system
```

You should see the operator pod running:

```
NAME                                READY   STATUS    RESTARTS   AGE
fetchfy-operator-7c9b4f8d8d-t2xjp   1/1     Running   0          1m
```

## Configuration Options

The operator can be configured through environment variables or by editing the operator deployment:

| Parameter                | Description                               | Default |
| ------------------------ | ----------------------------------------- | ------- |
| `METRICS_ADDR`           | Address to bind the metrics endpoint      | `:8080` |
| `HEALTH_PROBE_ADDR`      | Address to bind the health probe          | `:8081` |
| `ENABLE_LEADER_ELECTION` | Enable leader election for HA deployments | `false` |

### Customizing Resource Requests/Limits

For production deployments, you may want to adjust resource requests and limits:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

## Uninstalling

To remove the Fetchfy operator and its CRDs from your cluster:

```bash
# If installed via kubectl
kubectl delete -f https://github.com/fetchfy/fetchfy-operator/releases/latest/download/fetchfy-operator.yaml

# If installed via Helm
helm uninstall fetchfy-operator

# If installed from source
make uninstall
```

## Next Steps

Now that you have installed Fetchfy, you can:

- [Create your first Gateway](../guides/creating-gateway.md)
- [Deploy MCP Services](../guides/deploying-services.md)
- [Learn about the Gateway CRD](../api-reference/gateway-crd.md)
