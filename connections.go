package reevit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ConnectionsService handles communication with the connection related methods of the Reevit API.
type ConnectionsService service

// ConnectionRequest represents a request to create a connection.
type ConnectionRequest struct {
	ID           string                 `json:"id,omitempty"`
	Provider     string                 `json:"provider"`
	Mode         string                 `json:"mode"`
	Status       string                 `json:"status,omitempty"`
	Credentials  map[string]interface{} `json:"credentials"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
	RoutingHints *RoutingHints          `json:"routing_hints,omitempty"`
	Labels       []string               `json:"labels,omitempty"`
	WorkflowID   string                 `json:"workflow_id,omitempty"`
}

// Connection represents a connection object.
type Connection struct {
	ID                   string                  `json:"id"`
	Provider             string                  `json:"provider"`
	Mode                 string                  `json:"mode"`
	Status               string                  `json:"status"`
	Capabilities         map[string]interface{}  `json:"capabilities"`
	RoutingHints         *RoutingHints           `json:"routing_hints"`
	Labels               []string                `json:"labels"`
	WorkflowID           string                  `json:"workflow_id,omitempty"`
	FeeStructure         *ConnectionFeeStructure `json:"fee_structure,omitempty"`
	TotalCalls           int64                   `json:"total_calls,omitempty"`
	SuccessfulCalls      int64                   `json:"successful_calls,omitempty"`
	FailedCalls          int64                   `json:"failed_calls,omitempty"`
	TotalLatencyMs       int64                   `json:"total_latency_ms,omitempty"`
	LastSuccessAt        *time.Time              `json:"last_success_at,omitempty"`
	LastFailureAt        *time.Time              `json:"last_failure_at,omitempty"`
	HealthScore          *float64                `json:"health_score,omitempty"`
	HealthStatus         string                  `json:"health_status,omitempty"`
	LastValidatedAt      *time.Time              `json:"last_validated_at,omitempty"`
	LastValidationStatus string                  `json:"last_validation_status,omitempty"`
	LastValidationError  string                  `json:"last_validation_error,omitempty"`
	ValidationFailures   int                     `json:"validation_failures,omitempty"`
	NameMatchStatus      string                  `json:"name_match_status,omitempty"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
}

type ConnectionFee struct {
	Percentage float64 `json:"percentage,omitempty"`
	Fixed      int64   `json:"fixed,omitempty"`
}

type ConnectionFeeStructure struct {
	Percentage  float64                  `json:"percentage,omitempty"`
	Fixed       int64                    `json:"fixed,omitempty"`
	PerMethod   map[string]ConnectionFee `json:"per_method,omitempty"`
	PerCurrency map[string]ConnectionFee `json:"per_currency,omitempty"`
}

// ConnectionListOptions contains filters for connection listing.
type ConnectionListOptions struct {
	Limit    int
	Offset   int
	Provider string
	Mode     string
	Status   string
	Label    string
}

type ConnectionPagination struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type ConnectionListPage struct {
	Connections []Connection         `json:"connections"`
	Pagination  ConnectionPagination `json:"pagination"`
}

type ConnectionLabelStat struct {
	Label     string         `json:"label"`
	Total     int            `json:"total"`
	Providers map[string]int `json:"providers,omitempty"`
	Statuses  map[string]int `json:"statuses,omitempty"`
}

// ConnectionAuditEntry describes an audit trail item for a connection.
type ConnectionAuditEntry struct {
	ID        interface{}            `json:"id"`
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
	page, err := s.ListPage(ctx, options...)
	if err != nil {
		return nil, err
	}

	return page.Connections, nil
}

// ListPage returns connections together with the server's pagination metadata.
func (s *ConnectionsService) ListPage(ctx context.Context, options ...ConnectionListOptions) (*ConnectionListPage, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "provider", options[0].Provider)
		setString(values, "mode", options[0].Mode)
		setString(values, "status", options[0].Status)
		setString(values, "label", options[0].Label)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/connections", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	var page ConnectionListPage
	if err := json.Unmarshal(raw, &page); err == nil && page.Connections != nil {
		return &page, nil
	}

	var direct []Connection
	if err := json.Unmarshal(raw, &direct); err != nil {
		return nil, fmt.Errorf("reevit: decode connections response: %w", err)
	}
	page.Connections = direct
	page.Pagination.Total = int64(len(direct))
	page.Pagination.Limit = len(direct)
	if len(options) > 0 {
		page.Pagination.Offset = options[0].Offset
		if options[0].Limit > 0 {
			page.Pagination.Limit = options[0].Limit
		}
	}

	return &page, nil
}

// ListAll follows pagination until every connection matching the filters has
// been fetched. Limit and Offset in filters are ignored.
func (s *ConnectionsService) ListAll(ctx context.Context, filters ConnectionListOptions) ([]Connection, error) {
	all := make([]Connection, 0)
	offset := 0

	for {
		filters.Limit = 200
		filters.Offset = offset
		page, err := s.ListPage(ctx, filters)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Connections...)

		nextOffset := offset + len(page.Connections)
		if len(page.Connections) == 0 || int64(nextOffset) >= page.Pagination.Total {
			return all, nil
		}
		offset = nextOffset
	}
}

// Get retrieves a connection by ID.
//
// API Docs: GET /v1/connections/{id}
func (s *ConnectionsService) Get(ctx context.Context, connectionID string) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/connections/%s", url.PathEscape(connectionID)), nil)
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
	httpRequest, err := s.client.newRequest(http.MethodDelete, fmt.Sprintf("/v1/connections/%s", url.PathEscape(connectionID)), nil)
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
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/connections/%s/validate", url.PathEscape(connectionID)), map[string]interface{}{})
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

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath(fmt.Sprintf("/v1/connections/%s/audit", url.PathEscape(connectionID)), values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[ConnectionAuditEntry](raw, "audit")
}

// ListLabels returns the labels used by the merchant's connections.
func (s *ConnectionsService) ListLabels(ctx context.Context) ([]ConnectionLabelStat, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, "/v1/connections/labels", nil)
	if err != nil {
		return nil, err
	}
	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[ConnectionLabelStat](raw, "labels")
}

// UpdateLabels updates connection labels.
//
// API Docs: PATCH /v1/connections/{id}/labels
func (s *ConnectionsService) UpdateLabels(ctx context.Context, connectionID string, req *ConnectionLabelsUpdate, opts ...RequestOption) (*Connection, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/connections/%s/labels", url.PathEscape(connectionID)), req)
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
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/connections/%s/status", url.PathEscape(connectionID)), req)
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
		OK      bool `json:"ok"`
		Success bool `json:"success"`
	}
	if err := s.client.do(ctx, httpRequest, &result); err != nil {
		return false, err
	}

	return result.OK || result.Success, nil
}
