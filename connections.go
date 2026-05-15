package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
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
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ConnectionListOptions contains filters for connection listing.
type ConnectionListOptions struct {
	Limit    int
	Offset   int
	Provider string
	Mode     string
	Status   string
}

// ConnectionAuditEntry describes an audit trail item for a connection.
type ConnectionAuditEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	ActorID   string                 `json:"actor_id"`
	ActorType string                 `json:"actor_type"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// ConnectionLabelsUpdate updates the labels applied to a connection.
type ConnectionLabelsUpdate struct {
	Labels []string `json:"labels"`
}

// ConnectionStatusUpdate updates a connection status.
type ConnectionStatusUpdate struct {
	Status string `json:"status"`
}

// Create creates a new connection.
//
// API Docs: POST /v1/connections
func (s *ConnectionsService) Create(ctx context.Context, req *ConnectionRequest, opts ...RequestOption) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/connections", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
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
func (s *ConnectionsService) List(ctx context.Context, options ...ConnectionListOptions) ([]Connection, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "provider", options[0].Provider)
		setString(values, "mode", options[0].Mode)
		setString(values, "status", options[0].Status)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/connections", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[Connection](raw, "connections")
}

// Get retrieves a connection by ID.
//
// API Docs: GET /v1/connections/{id}
func (s *ConnectionsService) Get(ctx context.Context, connectionID string) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/connections/%s", connectionID), nil)
	if err != nil {
		return nil, err
	}

	var connection Connection
	if err := s.client.do(ctx, httpRequest, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

// Delete removes a connection.
//
// API Docs: DELETE /v1/connections/{id}
func (s *ConnectionsService) Delete(ctx context.Context, connectionID string, opts ...RequestOption) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, fmt.Sprintf("/v1/connections/%s", connectionID), nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	return s.client.do(ctx, httpRequest, nil)
}

// Validate validates a connection configuration.
//
// API Docs: POST /v1/connections/{id}/validate
func (s *ConnectionsService) Validate(ctx context.Context, connectionID string, opts ...RequestOption) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/connections/%s/validate", connectionID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var connection Connection
	if err := s.client.do(ctx, httpRequest, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

// ListAudit returns the audit history for a connection.
//
// API Docs: GET /v1/connections/{id}/audit
func (s *ConnectionsService) ListAudit(ctx context.Context, connectionID string, options ...ConnectionListOptions) ([]ConnectionAuditEntry, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath(fmt.Sprintf("/v1/connections/%s/audit", connectionID), values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[ConnectionAuditEntry](raw, "audit")
}

// UpdateLabels updates connection labels.
//
// API Docs: PATCH /v1/connections/{id}/labels
func (s *ConnectionsService) UpdateLabels(ctx context.Context, connectionID string, req *ConnectionLabelsUpdate, opts ...RequestOption) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/connections/%s/labels", connectionID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var connection Connection
	if err := s.client.do(ctx, httpRequest, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

// UpdateStatus updates connection status.
//
// API Docs: PATCH /v1/connections/{id}/status
func (s *ConnectionsService) UpdateStatus(ctx context.Context, connectionID string, req *ConnectionStatusUpdate, opts ...RequestOption) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/connections/%s/status", connectionID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var connection Connection
	if err := s.client.do(ctx, httpRequest, &connection); err != nil {
		return nil, err
	}

	return &connection, nil
}

// Test tests a connection.
//
// API Docs: POST /v1/connections/test
func (s *ConnectionsService) Test(ctx context.Context, req *ConnectionRequest, opts ...RequestOption) (bool, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/connections/test", req)
	if err != nil {
		return false, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.do(ctx, httpRequest, &result); err != nil {
		return false, err
	}

	return result.Success, nil
}
