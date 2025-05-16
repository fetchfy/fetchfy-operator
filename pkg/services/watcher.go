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

package services

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	fetchfyv1alpha1 "github.com/fetchfy/fetchfy-operator/api/v1alpha1"
	"github.com/fetchfy/fetchfy-operator/pkg/mcp"
)

const (
	// MCPEnabledLabel is the label that indicates a service is MCP-enabled
	MCPEnabledLabel = "mcp-enabled"

	// MCPTypeAnnotation is the annotation that specifies the MCP service type
	MCPTypeAnnotation = "mcp.fetchfy.ai/type"
)

// IsMCPEnabledService checks if a service is MCP-enabled
func IsMCPEnabledService(obj client.Object) bool {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return false
	}

	if svc.Labels == nil {
		return false
	}

	if val, exists := svc.Labels[MCPEnabledLabel]; exists && val == "true" {
		return true
	}

	return false
}

// ServiceWatcher watches for Kubernetes services that are MCP-enabled
// and registers/deregisters them with the MCP registry
type ServiceWatcher struct {
	client    client.Client
	log       logr.Logger
	registry  *mcp.Registry
	gateways  map[types.NamespacedName]*fetchfyv1alpha1.Gateway
	scheme    *runtime.Scheme
	predicate predicate.Predicate
}

// NewServiceWatcher creates a new service watcher
func NewServiceWatcher(
	client client.Client,
	registry *mcp.Registry,
	log logr.Logger,
	scheme *runtime.Scheme,
) *ServiceWatcher {

	sw := &ServiceWatcher{
		client:   client,
		registry: registry,
		log:      log.WithName("service-watcher"),
		gateways: make(map[types.NamespacedName]*fetchfyv1alpha1.Gateway),
		scheme:   scheme,
	}

	// Create a predicate that filters services based on labels
	sw.predicate = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return sw.isMCPEnabledService(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return sw.isMCPEnabledService(e.ObjectNew) || sw.isMCPEnabledService(e.ObjectOld)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return sw.isMCPEnabledService(e.Object)
		},
	}

	return sw
}

// IsMCPEnabledService checks if a service is MCP-enabled
func IsMCPEnabledService(obj client.Object) bool {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		return false
	}

	if svc.Labels == nil {
		return false
	}

	if val, exists := svc.Labels[MCPEnabledLabel]; exists && val == "true" {
		return true
	}

	return false
}

// isMCPEnabledService checks if a service is MCP-enabled (internal method)
func (sw *ServiceWatcher) isMCPEnabledService(obj client.Object) bool {
	return IsMCPEnabledService(obj)
}

// SetupWithManager sets up the service watcher with the manager
func (sw *ServiceWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		WithEventFilter(sw.predicate).
		Complete(sw)
}

// Reconcile handles service reconciliation
func (sw *ServiceWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := sw.log.WithValues("service", req.NamespacedName)

	// Fetch the service
	var service corev1.Service
	if err := sw.client.Get(ctx, req.NamespacedName, &service); err != nil {
		if errors.IsNotFound(err) {
			// Service deleted, deregister it
			if sw.registry.DeregisterService(ctx, req.NamespacedName) {
				log.Info("Deregistered service from MCP registry")
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch service")
		return ctrl.Result{}, err
	}

	// Check if the service is MCP-enabled
	if !sw.isMCPEnabledService(&service) {
		// Service is not MCP-enabled, deregister it if it was previously registered
		if sw.registry.DeregisterService(ctx, req.NamespacedName) {
			log.Info("Deregistered non-MCP service from registry")
		}
		return ctrl.Result{}, nil
	}

	// Determine service type from annotation
	serviceType := mcp.ServiceTypeTool // Default type
	if service.Annotations != nil {
		if typeStr, exists := service.Annotations[MCPTypeAnnotation]; exists {
			if typeStr == string(mcp.ServiceTypeAgent) {
				serviceType = mcp.ServiceTypeAgent
			}
		}
	}

	// Register the service
	_, err := sw.registry.RegisterService(ctx, &service, serviceType)
	if err != nil {
		log.Error(err, "Failed to register MCP service")
		return ctrl.Result{}, err
	}

	log.Info("Registered MCP service", "type", serviceType)

	// Update all gateway statuses
	sw.updateAllGatewayStatuses(ctx)

	return ctrl.Result{}, nil
}

// AddGateway adds a gateway to be tracked by the watcher
func (sw *ServiceWatcher) AddGateway(gateway *fetchfyv1alpha1.Gateway) {
	key := types.NamespacedName{
		Name:      gateway.Name,
		Namespace: gateway.Namespace,
	}
	sw.gateways[key] = gateway
	sw.log.Info("Added gateway to service watcher",
		"gateway", fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name))
}

// RemoveGateway removes a gateway from being tracked
func (sw *ServiceWatcher) RemoveGateway(name types.NamespacedName) {
	delete(sw.gateways, name)
	sw.log.Info("Removed gateway from service watcher",
		"gateway", fmt.Sprintf("%s/%s", name.Namespace, name.Name))
}

// GetMatchingServices returns all services that match the gateway's selector
func (sw *ServiceWatcher) GetMatchingServices(ctx context.Context, gateway *fetchfyv1alpha1.Gateway) ([]corev1.Service, error) {
	serviceList := &corev1.ServiceList{}

	selector, err := metav1.LabelSelectorAsSelector(&gateway.Spec.ServiceSelector)
	if err != nil {
		return nil, err
	}

	listOpts := &client.ListOptions{
		LabelSelector: selector,
	}

	if err := sw.client.List(ctx, serviceList, listOpts); err != nil {
		return nil, err
	}

	return serviceList.Items, nil
}

// updateAllGatewayStatuses updates the status of all tracked gateways
func (sw *ServiceWatcher) updateAllGatewayStatuses(ctx context.Context) {
	for key, _ := range sw.gateways {
		// Fetch the latest gateway
		var currentGateway fetchfyv1alpha1.Gateway
		if err := sw.client.Get(ctx, key, &currentGateway); err != nil {
			sw.log.Error(err, "Failed to fetch gateway for status update",
				"gateway", fmt.Sprintf("%s/%s", key.Namespace, key.Name))
			continue
		}

		// Update the gateway status with the registered services
		sw.registry.UpdateRegistryStatus(&currentGateway)

		// Update the gateway status
		if err := sw.client.Status().Update(ctx, &currentGateway); err != nil {
			sw.log.Error(err, "Failed to update gateway status",
				"gateway", fmt.Sprintf("%s/%s", key.Namespace, key.Name))
		}
	}
}
