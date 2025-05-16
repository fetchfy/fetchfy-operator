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

package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	fetchfyv1alpha1 "github.com/fetchfy/fetchfy-operator/api/v1alpha1"
)

// Server represents an MCP gateway server that handles connections and routes requests
// to the appropriate MCP services
type Server struct {
	registry      *Registry
	httpServer    *http.Server
	port          int32
	log           logr.Logger
	gatewayRef    types.NamespacedName
	enableTLS     bool
	tlsSecretName string
	mutex         sync.Mutex
	started       bool
}

// NewServer creates a new MCP gateway server
func NewServer(registry *Registry, log logr.Logger) *Server {
	return &Server{
		registry: registry,
		log:      log.WithName("mcp-server"),
		started:  false,
	}
}

// Configure configures the server with the gateway's settings
func (s *Server) Configure(gateway *fetchfyv1alpha1.Gateway) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.port = gateway.Spec.MCPPort
	s.enableTLS = gateway.Spec.EnableTLS
	s.tlsSecretName = gateway.Spec.TLSSecretRef
	s.gatewayRef = types.NamespacedName{
		Name:      gateway.Name,
		Namespace: gateway.Namespace,
	}

	s.log.Info("Configured MCP server",
		"port", s.port,
		"enableTLS", s.enableTLS,
		"gateway", fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name))
}

// Start starts the MCP gateway server
func (s *Server) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	mux := http.NewServeMux()

	// Root handler for health checks
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Fetchfy MCP Gateway: OK"))
	})

	// MCP routes handler
	mux.HandleFunc("/mcp/", s.handleMCPRequest)

	// API endpoints for MCP management
	mux.HandleFunc("/api/services", s.handleListServices)

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.log.Info("Starting MCP gateway server", "address", addr)

	go func() {
		var err error
		if s.enableTLS && s.tlsSecretName != "" {
			// TODO: Implement TLS with cert from secret
			s.log.Error(fmt.Errorf("TLS support not implemented yet"), "Failed to start server with TLS")
			err = s.httpServer.ListenAndServe()
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.log.Error(err, "MCP gateway server failed")
		}
	}()

	s.started = true
	return nil
}

// Stop stops the MCP gateway server
func (s *Server) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started || s.httpServer == nil {
		return nil
	}

	s.log.Info("Stopping MCP gateway server")
	err := s.httpServer.Shutdown(ctx)
	s.started = false
	return err
}

// IsRunning returns true if the server is running
func (s *Server) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.started
}

// handleMCPRequest handles MCP protocol requests and routes them to the appropriate service
func (s *Server) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// This is a simplified implementation. In a real-world scenario,
	// this would handle routing to the actual MCP services based on the endpoint path

	// For now, just return a simple response to indicate the request was received
	s.log.Info("Received MCP request", "path", r.URL.Path, "method", r.Method)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"MCP request received"}`))
}

// handleListServices returns a list of registered MCP services
func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := s.registry.ListServices()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response with service list
	w.Write([]byte(fmt.Sprintf(`{"services":%d}`, len(services))))
}
