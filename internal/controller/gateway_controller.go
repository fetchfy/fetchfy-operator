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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fetchfyv1alpha1 "github.com/fetchfy/fetchfy-operator/api/v1alpha1"
	"github.com/fetchfy/fetchfy-operator/pkg/mcp"
	"github.com/fetchfy/fetchfy-operator/pkg/services"
)

const (
	gatewayFinalizer = "fetchfy.ai/finalizer"

	// Condition types
	conditionTypeReady     = "Ready"
	conditionTypeAvailable = "Available"

	// Condition reasons
	reasonReady       = "GatewayReady"
	reasonConfigured  = "GatewayConfigured"
	reasonNotReady    = "GatewayNotReady"
	reasonServerError = "ServerError"
	reasonConfigError = "ConfigurationError"
)

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Recorder       record.EventRecorder
	MCPRegistry    *mcp.Registry
	ServiceWatcher *services.ServiceWatcher
	MCPServers     map[types.NamespacedName]*mcp.Server
	Log            logr.Logger
}

// +kubebuilder:rbac:groups=fetchfy.fetchfy.ai,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fetchfy.fetchfy.ai,resources=gateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fetchfy.fetchfy.ai,resources=gateways/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile handles reconciliation of Gateway resources
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("gateway", req.NamespacedName)

	// Fetch the Gateway instance
	gateway := &fetchfyv1alpha1.Gateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if errors.IsNotFound(err) {
			// Gateway might have been deleted, clean up
			r.cleanupGateway(ctx, req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch Gateway")
		return ctrl.Result{}, err
	}

	// Initialize status conditions if they don't exist
	if gateway.Status.Conditions == nil {
		gateway.Status.Conditions = []metav1.Condition{}
	}

	// Check if the gateway is being deleted
	if !gateway.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, gateway)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(gateway, gatewayFinalizer) {
		controllerutil.AddFinalizer(gateway, gatewayFinalizer)
		if err := r.Update(ctx, gateway); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle MCP server setup/configuration
	if err := r.ensureMCPServer(ctx, gateway); err != nil {
		log.Error(err, "Failed to ensure MCP server")

		// Update gateway status to reflect the error
		r.updateGatewayCondition(ctx, gateway, conditionTypeReady, metav1.ConditionFalse, reasonServerError, err.Error())

		return ctrl.Result{RequeueAfter: time.Second * 30}, err
	}

	// Configure service watcher
	r.ServiceWatcher.AddGateway(gateway)

	// Get matching services for this gateway
	matchingServices, err := r.ServiceWatcher.GetMatchingServices(ctx, gateway)
	if err != nil {
		log.Error(err, "Failed to list matching services")
		r.updateGatewayCondition(ctx, gateway, conditionTypeReady, metav1.ConditionFalse, reasonConfigError,
			"Failed to list matching services: "+err.Error())
		return ctrl.Result{RequeueAfter: time.Second * 30}, err
	}

	// Register all matching services
	for _, svc := range matchingServices {
		// Check if the service has the MCP enabled label
		if svc.Labels != nil && svc.Labels[services.MCPEnabledLabel] == "true" {
			serviceType := mcp.ServiceTypeTool
			if svc.Annotations != nil && svc.Annotations[services.MCPTypeAnnotation] == string(mcp.ServiceTypeAgent) {
				serviceType = mcp.ServiceTypeAgent
			}
			if _, err := r.MCPRegistry.RegisterService(ctx, &svc, serviceType); err != nil {
				log.Error(err, "Failed to register service", "service", fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
			}
		}
	}

	// Update gateway status
	r.MCPRegistry.UpdateRegistryStatus(gateway)

	// Set address field in status
	gateway.Status.Address = fmt.Sprintf(":%d", gateway.Spec.MCPPort)

	// Update gateway status to Ready
	r.updateGatewayCondition(ctx, gateway, conditionTypeReady, metav1.ConditionTrue, reasonReady,
		fmt.Sprintf("Gateway is ready with %d services", len(gateway.Status.MCPServices)))

	r.updateGatewayCondition(ctx, gateway, conditionTypeAvailable, metav1.ConditionTrue, reasonConfigured,
		"Gateway is available")

	if err := r.Status().Update(ctx, gateway); err != nil {
		log.Error(err, "Failed to update Gateway status")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// Requeue periodically to ensure gateway status stays up to date
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// handleDeletion handles the deletion of a Gateway resource
func (r *GatewayReconciler) handleDeletion(ctx context.Context, gateway *fetchfyv1alpha1.Gateway) (ctrl.Result, error) {
	log := r.Log.WithValues("gateway", types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace})
	log.Info("Handling deletion of Gateway")

	// Clean up resources
	r.cleanupGateway(ctx, types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace})

	// Check if the finalizer exists
	if controllerutil.ContainsFinalizer(gateway, gatewayFinalizer) {
		// Remove the finalizer
		controllerutil.RemoveFinalizer(gateway, gatewayFinalizer)
		if err := r.Update(ctx, gateway); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// cleanupGateway cleans up resources associated with a gateway
func (r *GatewayReconciler) cleanupGateway(ctx context.Context, gatewayName types.NamespacedName) {
	log := r.Log.WithValues("gateway", gatewayName)
	log.Info("Cleaning up resources for gateway")

	// Stop and remove MCP server
	if server, exists := r.MCPServers[gatewayName]; exists {
		if err := server.Stop(ctx); err != nil {
			log.Error(err, "Failed to stop MCP server")
		}
		delete(r.MCPServers, gatewayName)
	}

	// Remove gateway from service watcher
	r.ServiceWatcher.RemoveGateway(gatewayName)
}

// ensureMCPServer ensures that an MCP server is running for the gateway
func (r *GatewayReconciler) ensureMCPServer(ctx context.Context, gateway *fetchfyv1alpha1.Gateway) error {
	gatewayName := types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}

	// Check if server exists
	server, exists := r.MCPServers[gatewayName]
	if !exists {
		// Create new server
		server = mcp.NewServer(r.MCPRegistry, r.Log)
		r.MCPServers[gatewayName] = server
	}

	// Configure server
	server.Configure(gateway)

	// Start server if not running
	if !server.IsRunning() {
		if err := server.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// updateGatewayCondition updates a condition in the gateway status
func (r *GatewayReconciler) updateGatewayCondition(
	ctx context.Context,
	gateway *fetchfyv1alpha1.Gateway,
	conditionType string,
	status metav1.ConditionStatus,
	reason, message string,
) {
	// Find existing condition if it exists
	for i, condition := range gateway.Status.Conditions {
		if condition.Type == conditionType {
			// Don't update if nothing changed
			if condition.Status == status && condition.Reason == reason && condition.Message == message {
				return
			}

			// Update existing condition
			gateway.Status.Conditions[i] = metav1.Condition{
				Type:               conditionType,
				Status:             status,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Reason:             reason,
				Message:            message,
				ObservedGeneration: gateway.Generation,
			}
			return
		}
	}

	// Add new condition
	gateway.Status.Conditions = append(gateway.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
		ObservedGeneration: gateway.Generation,
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize MCPServers map if not set
	if r.MCPServers == nil {
		r.MCPServers = make(map[types.NamespacedName]*mcp.Server)
	}

	// Initialize logger
	logger := logf.Log.WithName("gateway-controller")
	if r.Log.GetSink() == nil {
		r.Log = logger
	}

	// Create MCP Registry if not provided
	if r.MCPRegistry == nil {
		r.MCPRegistry = mcp.NewRegistry(r.Log)
	}

	// Create Service Watcher if not provided
	if r.ServiceWatcher == nil {
		r.ServiceWatcher = services.NewServiceWatcher(r.Client, r.MCPRegistry, r.Log, r.Scheme)
		if err := r.ServiceWatcher.SetupWithManager(mgr); err != nil {
			return err
		}
	}

	// Setup event recorder
	r.Recorder = mgr.GetEventRecorderFor("gateway-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&fetchfyv1alpha1.Gateway{}).
		Complete(r)
}
