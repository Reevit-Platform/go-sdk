package reevit

import (
	"context"
	"net/http"
)

// ConnectionsService handles communication with the connection related methods of the Reevit API.
type ConnectionsService service

// ConnectionRequest represents a request to create a connection.
type ConnectionRequest struct {
	Provider     string                 `json:"provider"`
	Mode         string                 `json:"mode"`
	Credentials  map[string]interface{} `json:"credentials"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
	RoutingHints *RoutingHints          `json:"routing_hints,omitempty"`
	Labels       []string               `json:"labels,omitempty"`
}

// Connection represents a connection object.
type Connection struct {
	ID           string                 `json:"id"`
	Provider     string                 `json:"provider"`
	Mode         string                 `json:"mode"`
	Status       string                 `json:"status"`
	Capabilities map[string]interface{} `json:"capabilities"`
	RoutingHints *RoutingHints          `json:"routing_hints"`
	Labels       []string               `json:"labels"`
}

// Create creates a new connection.
//
// API Docs: POST /v1/connections
func (s *ConnectionsService) Create(ctx context.Context, req *ConnectionRequest) (*Connection, error) {
	path := "v1/connections"

	httpRequest, err := s.client.newRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var connection Connection
	if err := s.client.do(ctx, httpRequest, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

// List returns a list of connections.
//
// API Docs: GET /v1/connections
func (s *ConnectionsService) List(ctx context.Context) ([]Connection, error) {
	path := "v1/connections"

	httpRequest, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var connections []Connection
	if err := s.client.do(ctx, httpRequest, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

// Test tests a connection.
//
// API Docs: POST /v1/connections/test
func (s *ConnectionsService) Test(ctx context.Context, req *ConnectionRequest) (bool, error) {
	path := "v1/connections/test"

	httpRequest, err := s.client.newRequest(http.MethodPost, path, req)
	if err != nil {
		return false, err
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.do(ctx, httpRequest, &result); err != nil {
		return false, err
	}

	return result.Success, nil
}
