# Creating a Gateway

This guide explains how to create and configure a Fetchfy Gateway to manage MCP services in your Kubernetes cluster.

## Basic Gateway Creation

Creating a basic Gateway is a simple process that requires just a few configuration parameters.

### Minimal Gateway

Create a file named `gateway.yaml` with the following content:

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

Apply it to your cluster:

```bash
kubectl apply -f gateway.yaml
```

This creates a basic Gateway that:

- Listens on port 8080 for MCP traffic
- Registers services with the label `mcp-enabled: "true"`

### Verifying Gateway Status

Check the status of your Gateway:

```bash
kubectl get gateway fetchfy-gateway
```

The output should look similar to:

```
NAME             PORT   SERVICES   ADDRESS   AGE
fetchfy-gateway  8080   2          :8080     2m
```

For more detailed information:

```bash
kubectl describe gateway fetchfy-gateway
```

## Advanced Gateway Configuration

For more complex scenarios, you can customize the Gateway configuration.

### Gateway with TLS

For secure MCP communications, enable TLS:

1. First, create a TLS secret with your certificate and key:

```bash
kubectl create secret tls mcp-tls-secret \
  --cert=path/to/cert.crt \
  --key=path/to/private.key
```

2. Create the Gateway with TLS enabled:

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

### Multiple Gateways

You can run multiple Gateways in the same cluster, each with different configurations:

```yaml
---
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: internal-gateway
  namespace: default
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
      visibility: "internal"
---
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: external-gateway
  namespace: default
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
      visibility: "external"
  enableTls: true
  tlsSecretRef: "external-tls-secret"
```

### Advanced Service Selection

You can use more complex selectors to target specific services:

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: advanced-gateway
spec:
  mcpPort: 8080
  serviceSelector:
    matchExpressions:
      - key: mcp-enabled
        operator: In
        values: ["true", "yes"]
      - key: environment
        operator: NotIn
        values: ["test", "staging"]
```

This Gateway will register services that:

- Have the label `mcp-enabled` set to either "true" or "yes"
- Don't have the `environment` label set to "test" or "staging"

## Gateway per Namespace

For multi-tenant environments, you might want to create a Gateway in each namespace:

```yaml
apiVersion: fetchfy.ai/v1alpha1
kind: Gateway
metadata:
  name: team-a-gateway
  namespace: team-a
spec:
  mcpPort: 8080
  serviceSelector:
    matchLabels:
      mcp-enabled: "true"
```

## Exposing the Gateway

### Internal Access (within cluster)

By default, the Gateway is available within the cluster at:

```
<gateway-name>.<namespace>.svc.cluster.local:<port>
```

For example:

```
fetchfy-gateway.default.svc.cluster.local:8080
```

### External Access

To expose the Gateway outside the cluster, create a Kubernetes Service of type LoadBalancer:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway-lb
spec:
  selector:
    app.kubernetes.io/name: fetchfy-operator
    app.kubernetes.io/instance: fetchfy
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
  type: LoadBalancer
```

Or use an Ingress controller:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gateway-ingress
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
spec:
  rules:
    - host: mcp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: fetchfy-gateway
                port:
                  number: 8080
  tls:
    - hosts:
        - mcp.example.com
      secretName: mcp-tls-secret
```

## Monitoring Gateway Status

### Viewing Registered Services

To see which services are registered with the Gateway:

```bash
kubectl get gateway fetchfy-gateway -o jsonpath="{.status.mcpServices}" | jq .
```

### Checking Gateway Conditions

To check the Gateway's conditions:

```bash
kubectl get gateway fetchfy-gateway -o jsonpath="{.status.conditions}" | jq .
```

## Gateway Lifecycle Management

### Updating a Gateway

To update a Gateway configuration, simply edit and reapply the YAML:

```bash
kubectl edit gateway fetchfy-gateway
```

Or:

```bash
# After modifying gateway.yaml
kubectl apply -f gateway.yaml
```

### Deleting a Gateway

To remove a Gateway:

```bash
kubectl delete gateway fetchfy-gateway
```

This will:

- Stop the MCP Gateway server
- Deregister all associated services
- Remove the Gateway resource

## Best Practices

1. **Port Selection**: Choose a port that doesn't conflict with other services
2. **TLS**: Use TLS for production environments
3. **Resource Limits**: If running multiple Gateways, consider setting resource limits
4. **High Availability**: For production, deploy the operator with leader election enabled
5. **Monitoring**: Set up monitoring for the Gateway and operator

## Troubleshooting

### Gateway Not Ready

If the Gateway is not reaching the Ready status:

1. Check for errors in the Gateway status:

   ```bash
   kubectl describe gateway fetchfy-gateway
   ```

2. Check the operator logs:
   ```bash
   kubectl logs -n fetchfy-system deployment/fetchfy-operator-controller-manager
   ```

### Address Conflicts

If you see port binding errors in the logs, make sure the chosen port is not already in use by another service in the cluster.

### TLS Issues

If TLS is not working:

1. Ensure the secret exists and is correctly referenced
2. Check that the certificate and key are valid
3. Verify the operator has permissions to read the secret

## Next Steps

After creating a Gateway, you can:

- [Deploy MCP Services](deploying-services.md)
- [Set up monitoring](monitoring.md)
- [Configure security](security.md)
