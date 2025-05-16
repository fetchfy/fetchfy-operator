#!/bin/bash
# Demo script to show Fetchfy MCP Gateway in action

# Set variables
NAMESPACE="fetchfy-demo"

# Create namespace if it doesn't exist
echo "Creating namespace: $NAMESPACE"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Deploy the Gateway
echo "Deploying the MCP Gateway..."
kubectl apply -f config/samples/fetchfy_v1alpha1_gateway.yaml -n $NAMESPACE

# Wait for Gateway to be ready
echo "Waiting for Gateway to be ready..."
kubectl wait --for=condition=Ready gateway/fetchfy-gateway -n $NAMESPACE --timeout=60s

# Deploy a tool service
echo "Deploying MCP Tool service..."
kubectl apply -f config/samples/mcp_tool_example.yaml -n $NAMESPACE

# Deploy an agent service
echo "Deploying MCP Agent service..."
kubectl apply -f config/samples/mcp_agent_example.yaml -n $NAMESPACE

# Wait for services to be registered
echo "Waiting for services to be registered..."
sleep 5

# Check the Gateway status to verify services are registered
echo "Checking Gateway status..."
kubectl get gateway fetchfy-gateway -n $NAMESPACE -o jsonpath="{.status.mcpServices}" | jq .

echo ""
echo "Demo setup complete! MCP Gateway is running with registered services."
echo "Access the MCP Gateway at: http://fetchfy-gateway.$NAMESPACE:8080"
