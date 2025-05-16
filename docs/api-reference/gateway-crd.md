# Gateway CRD Reference

This page provides a detailed reference for the Gateway Custom Resource Definition (CRD), which is the primary API resource used to configure Fetchfy MCP Gateways.

## Gateway Resource Definition

The Gateway CRD is defined in the `fetchfy.ai/v1alpha1` API group.

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: example-gateway
spec:
  # ... spec fields ...
status:
  # ... status fields ...
```

## Spec Fields

| Field             | Type                                                                                                        | Required | Description                                                                                                   |
| ----------------- | ----------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------- |
| `mcpPort`         | integer                                                                                                     | Yes      | The port where the MCP gateway will be exposed. Valid range: 1-65535.                                         |
| `serviceSelector` | [LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta) | Yes      | Label selector used to identify MCP-enabled services to be registered with the gateway.                       |
| `enableTls`       | boolean                                                                                                     | No       | Whether to enable TLS for secure MCP communication. Default: `false`.                                         |
| `tlsSecretRef`    | string                                                                                                      | No       | Reference to the Kubernetes secret containing the TLS certificate and key. Required if `enableTls` is `true`. |

### LabelSelector

The `serviceSelector` field uses Kubernetes LabelSelector format:

```yaml
serviceSelector:
  matchLabels:
    key1: value1
    key2: value2
  matchExpressions:
    - key: key3
      operator: In
      values: ["val1", "val2"]
```

## Status Fields

The Gateway controller populates the following status fields:

| Field         | Type             | Description                                                                                                                                        |
| ------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `address`     | string           | The address where the MCP gateway is available.                                                                                                    |
| `mcpServices` | []MCPServiceInfo | List of MCP services registered with the gateway.                                                                                                  |
| `conditions`  | []Condition      | Standard Kubernetes [conditions](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions) reflecting the gateway's state. |

### MCPServiceInfo

Each entry in the `mcpServices` array contains the following fields:

| Field         | Type               | Description                                                              |
| ------------- | ------------------ | ------------------------------------------------------------------------ |
| `name`        | string             | Name of the registered service.                                          |
| `namespace`   | string             | Namespace of the registered service.                                     |
| `type`        | string             | Type of MCP service: "tool" or "agent".                                  |
| `endpoint`    | string             | The endpoint path for the service.                                       |
| `status`      | string             | Current status of the service: "Available", "Pending", or "Unavailable". |
| `lastUpdated` | string (timestamp) | When the service was last updated.                                       |

### Conditions

The standard conditions used by the Gateway controller:

| Type        | Status         | Reason                                   | Description                                                    |
| ----------- | -------------- | ---------------------------------------- | -------------------------------------------------------------- |
| `Ready`     | `True`/`False` | `GatewayReady`/`GatewayNotReady`         | Indicates if the gateway is operational.                       |
| `Available` | `True`/`False` | `GatewayConfigured`/`ConfigurationError` | Indicates if the gateway is properly configured and available. |

## Examples

### Basic Gateway

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: basic-gateway
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
```

### Gateway with TLS

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

### Gateway with Complex Service Selection

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: advanced-gateway
spec:
  mcpPort: 9000
  serviceSelector:
    matchExpressions:
      - key: mcp-enabled
        operator: In
        values: ["true", "yes"]
      - key: environment
        operator: NotIn
        values: ["test", "staging"]
```

## Status Example

Here's an example of what the status might look like for a running Gateway:

```yaml
status:
  address: ":8080"
  mcpServices:
    - name: calculator-tool
      namespace: default
      type: tool
      endpoint: "/mcp/tools/calculator"
      status: Available
      lastUpdated: "2025-05-16T15:23:42Z"
    - name: assistant-agent
      namespace: ai-services
      type: agent
      endpoint: "/mcp/agents/assistant"
      status: Available
      lastUpdated: "2025-05-16T12:05:18Z"
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-05-16T10:30:00Z"
      reason: GatewayReady
      message: "Gateway is ready with 2 services"
    - type: Available
      status: "True"
      lastTransitionTime: "2025-05-16T10:30:00Z"
      reason: GatewayConfigured
      message: "Gateway is available"
```

## API Validation

The Gateway CRD includes validation to ensure that:

1. `mcpPort` is within the valid range (1-65535)
2. `serviceSelector` is properly formatted
3. `tlsSecretRef` is provided when `enableTls` is `true`

These validations are enforced through OpenAPI v3 schema in the CRD.

## Field Defaulting

The following defaults are applied:

- `enableTls`: `false`
- `mcpPort`: No default (required field)
- `serviceSelector`: No default (required field)

## Versioning and Compatibility

The current CRD version is `v1alpha1`, indicating that it's still in alpha stage and might change in future releases. API compatibility will be maintained according to Kubernetes API versioning guidelines once the API reaches a stable version.
