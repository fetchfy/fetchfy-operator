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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPServiceInfo provides information about a registered MCP service
type MCPServiceInfo struct {
	// Name is the name of the service
	Name string `json:"name"`

	// Namespace is the namespace where the service is deployed
	Namespace string `json:"namespace"`

	// Type indicates whether this is a tool or agent
	Type string `json:"type"`

	// Endpoint is the MCP endpoint for this service
	Endpoint string `json:"endpoint"`

	// Status indicates the current status of this service
	Status string `json:"status"`

	// LastUpdated is the timestamp of the last update
	LastUpdated metav1.Time `json:"lastUpdated"`
}

// GatewaySpec defines the desired state of Gateway.
type GatewaySpec struct {
	// MCPPort defines the port where the MCP gateway is available
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	MCPPort int32 `json:"mcpPort"`

	// ServiceSelector defines the label selector to identify MCP-enabled services
	// +kubebuilder:validation:Required
	ServiceSelector metav1.LabelSelector `json:"serviceSelector"`

	// EnableTLS indicates whether TLS should be enabled for the MCP gateway
	// +optional
	EnableTLS bool `json:"enableTls,omitempty"`

	// TLSSecretRef refers to the secret containing the TLS certificate and private key
	// +optional
	TLSSecretRef string `json:"tlsSecretRef,omitempty"`
}

// GatewayStatus defines the observed state of Gateway.
type GatewayStatus struct {
	// Conditions represent the latest available observations of Gateway's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// MCPServices contains information about registered MCP services
	// +optional
	MCPServices []MCPServiceInfo `json:"mcpServices,omitempty"`

	// Address where the MCP gateway is available
	// +optional
	Address string `json:"address,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=mcpgw
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.mcpPort",description="MCP Gateway port"
// +kubebuilder:printcolumn:name="Services",type="integer",JSONPath=".status.mcpServices",description="Number of registered services"
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".status.address",description="Gateway address"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Gateway is the Schema for the gateways API.
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec,omitempty"`
	Status GatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateway.
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
