/*
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
*/

// Package mcp provides utilities for Model Context Protocol integration
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	fetchfyv1alpha1 "github.com/fetchfy/fetchfy-operator/api/v1alpha1"
)

// ServiceType represents the type of MCP service
type ServiceType string

const (
	// ServiceTypeTool represents a tool service
	ServiceTypeTool ServiceType = "tool"

	// ServiceTypeAgent represents an agent service
	ServiceTypeAgent ServiceType = "agent"
)

// ServiceStatus represents the status of an MCP service
type ServiceStatus string

const (
	// ServiceStatusAvailable means the service is available and registered
	ServiceStatusAvailable ServiceStatus = "Available"

	// ServiceStatusPending means the service is registered but not yet ready
	ServiceStatusPending ServiceStatus = "Pending"

	// ServiceStatusUnavailable means the service is registered but cannot be reached
	ServiceStatusUnavailable ServiceStatus = "Unavailable"
)

// MCPService represents an MCP service registered with the gateway
type MCPService struct {
	Name      string
	Namespace string
	Type      ServiceType
	Endpoint  string
	Status    ServiceStatus
	Service   *corev1.Service
	UpdatedAt time.Time
}

// Registry maintains a registry of MCP services
type Registry struct {
	services map[types.NamespacedName]*MCPService
	mutex    sync.RWMutex
	log      logr.Logger
}

// NewRegistry creates a new MCP service registry
func NewRegistry(log logr.Logger) *Registry {
	return &Registry{
		services: make(map[types.NamespacedName]*MCPService),
		log:      log.WithName("mcp-registry"),
	}
}

// RegisterService adds or updates a service in the registry
func (r *Registry) RegisterService(ctx context.Context, svc *corev1.Service, serviceType ServiceType) (*MCPService, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}

	// Check if the service has the required annotations
	if svc.Annotations == nil {
		return nil, fmt.Errorf("service %s/%s has no annotations", svc.Namespace, svc.Name)
	}

	// Extract endpoint from annotations or generate one
	endpoint, ok := svc.Annotations["mcp.fetchfy.ai/endpoint"]
	if !ok {
		// Generate default endpoint based on service name
		endpoint = fmt.Sprintf("/mcp/%s/%s", svc.Namespace, svc.Name)
	}

	status := ServiceStatusPending
	if svc.Spec.Type == corev1.ServiceTypeClusterIP && len(svc.Spec.Ports) > 0 {
		status = ServiceStatusAvailable
	}

	mcpService := &MCPService{
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Type:      serviceType,
		Endpoint:  endpoint,
		Status:    status,
		Service:   svc.DeepCopy(),
		UpdatedAt: time.Now(),
	}

	r.services[key] = mcpService
	r.log.Info("Registered MCP service", "name", svc.Name, "namespace", svc.Namespace, "type", serviceType)

	return mcpService, nil
}

// DeregisterService removes a service from the registry
func (r *Registry) DeregisterService(ctx context.Context, name types.NamespacedName) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.services[name]; exists {
		delete(r.services, name)
		r.log.Info("Deregistered MCP service", "name", name.Name, "namespace", name.Namespace)
		return true
	}

	return false
}

// GetService returns a service from the registry
func (r *Registry) GetService(name types.NamespacedName) (*MCPService, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	svc, exists := r.services[name]
	return svc, exists
}

// ListServices returns all services in the registry
func (r *Registry) ListServices() []*MCPService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	services := make([]*MCPService, 0, len(r.services))
	for _, svc := range r.services {
		services = append(services, svc)
	}

	return services
}

// UpdateRegistryStatus updates the Gateway's status with current services
func (r *Registry) UpdateRegistryStatus(gateway *fetchfyv1alpha1.Gateway) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	mcpServices := make([]fetchfyv1alpha1.MCPServiceInfo, 0, len(r.services))
	for _, svc := range r.services {
		mcpServices = append(mcpServices, fetchfyv1alpha1.MCPServiceInfo{
			Name:        svc.Name,
			Namespace:   svc.Namespace,
			Type:        string(svc.Type),
			Endpoint:    svc.Endpoint,
			Status:      string(svc.Status),
			LastUpdated: metav1.NewTime(svc.UpdatedAt),
		})
	}

	gateway.Status.MCPServices = mcpServices
}
