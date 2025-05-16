---
title: Security Guide
description: Best practices for securing your Fetchfy MCP Gateway deployment
---

# Security Guide

This guide covers the security considerations and best practices when deploying the Fetchfy MCP Gateway Operator in your Kubernetes cluster.

## TLS Encryption

### Enabling TLS for the MCP Gateway

To secure communication between clients and the MCP Gateway, enable TLS by configuring the Gateway resource:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: secure-mcp-gateway
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  tls:
    enabled: true
    secretName: mcp-gateway-tls
```

### Creating TLS Certificates

You can generate certificates using cert-manager or manually create a Kubernetes Secret:

=== "Using cert-manager"

    ```yaml
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: mcp-gateway-cert
      namespace: fetchfy-system
    spec:
      secretName: mcp-gateway-tls
      duration: 2160h # 90 days
      renewBefore: 360h # 15 days
      subject:
        organizations:
          - Fetchfy
      isCA: false
      privateKey:
        algorithm: RSA
        encoding: PKCS1
        size: 2048
      usages:
        - server auth
      dnsNames:
        - mcp-gateway.fetchfy-system.svc
        - mcp-gateway.fetchfy-system.svc.cluster.local
      issuerRef:
        name: cluster-issuer
        kind: ClusterIssuer
    ```

=== "Manual Secret Creation"

    ```bash
    # Generate a private key
    openssl genrsa -out tls.key 2048

    # Generate a certificate signing request
    openssl req -new -key tls.key -out tls.csr -subj "/CN=mcp-gateway.fetchfy-system.svc"

    # Generate a self-signed certificate
    openssl x509 -req -in tls.csr -signkey tls.key -out tls.crt -days 365

    # Create the Kubernetes Secret
    kubectl create secret tls mcp-gateway-tls \
      --cert=tls.crt \
      --key=tls.key \
      -n fetchfy-system
    ```

### Certificate Rotation

The Fetchfy MCP Gateway Operator supports automatic certificate rotation. Configure the rotation interval in the Gateway resource:

```yaml
spec:
  tls:
    enabled: true
    secretName: mcp-gateway-tls
    rotationInterval: 48 # hours
```

## Authentication

### API Authentication

To restrict access to the MCP Gateway, configure authentication:

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: secure-mcp-gateway
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  authentication:
    enabled: true
    type: bearer
    secretName: mcp-gateway-auth
```

Create an authentication secret:

```bash
kubectl create secret generic mcp-gateway-auth \
  --from-literal=api-key="your-secure-api-key" \
  -n fetchfy-system
```

### Service-to-Service Authentication

For service-to-service authentication, the operator can use mutual TLS (mTLS):

```yaml
apiVersion: fetchfy.io/v1alpha1
kind: Gateway
metadata:
  name: secure-mcp-gateway
spec:
  mcpPort: 8443
  serviceSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  tls:
    enabled: true
    secretName: mcp-gateway-tls
    mutual: true
    caSecretName: mcp-gateway-ca
```

## Network Policies

### Restricting Gateway Access

Implement Kubernetes Network Policies to restrict which pods can access the MCP Gateway:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-gateway-network-policy
  namespace: fetchfy-system
spec:
  podSelector:
    matchLabels:
      app: mcp-gateway
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ai-models
        - podSelector:
            matchLabels:
              app: model-server
      ports:
        - protocol: TCP
          port: 8443
```

### Isolating MCP Tools

Restrict which pods can communicate with your MCP tools:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-tool-network-policy
  namespace: fetchfy-system
spec:
  podSelector:
    matchLabels:
      mcp.fetchfy.io/type: tool
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: mcp-gateway
      ports:
        - protocol: TCP
```

## RBAC Configuration

### Operator RBAC

The Fetchfy MCP Gateway Operator requires specific RBAC permissions to function. The Helm chart or manifests create these automatically, but you should review them:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: fetchfy-manager-role
rules:
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["fetchfy.io"]
    resources: ["gateways"]
    verbs: ["create", "delete", "get", "list", "patch", "update", "watch"]
  - apiGroups: ["fetchfy.io"]
    resources: ["gateways/status"]
    verbs: ["get", "patch", "update"]
```

### Least Privilege Principle

Follow the principle of least privilege by restricting the operator's permissions to only what is necessary:

- Use namespace-scoped roles when possible
- Avoid using cluster-admin privileges
- Define explicit resource permissions
- Restrict service account usage

## Pod Security

### Security Context

Configure the security context for the operator and gateway pods:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  seccompProfile:
    type: RuntimeDefault
```

### Pod Security Standards

Apply Kubernetes Pod Security Standards:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: fetchfy-system
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Monitoring and Alerting

### Security Monitoring

Set up monitoring and alerts for potential security issues:

- Watch for unauthorized access attempts
- Monitor certificate expiration
- Alert on configuration changes
- Track resource usage anomalies

The operator exposes Prometheus metrics at `/metrics` that can be used for monitoring.

## Auditing

Enable Kubernetes audit logs to track access to the Fetchfy API resources:

```yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: RequestResponse
    resources:
      - group: "fetchfy.io"
        resources: ["gateways"]
```

## Security Best Practices

1. **Regular Updates**: Keep the Fetchfy Operator and all dependencies up to date
2. **Secrets Management**: Use a solution like Vault or Sealed Secrets for managing secrets
3. **Image Security**: Scan container images for vulnerabilities
4. **Namespace Isolation**: Deploy the operator and MCP services in dedicated namespaces
5. **Resource Limits**: Set appropriate resource limits to prevent DoS attacks
6. **Disable Unused Features**: Turn off any features you don't need

## Next Steps

- Review the [Configuration Guide](../getting-started/configuration.md) for additional security settings
- Learn about [Monitoring](./monitoring.md) capabilities
- Implement proper [Gateway Configuration](../guides/creating-gateway.md)
