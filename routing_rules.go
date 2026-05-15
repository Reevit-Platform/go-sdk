package reevit

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// RoutingRulesService handles routing rule related methods of the Reevit API.
type RoutingRulesService service

// RoutingRule represents a routing rule resource.
type RoutingRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Priority   int                    `json:"priority"`
	Conditions map[string]interface{} `json:"conditions"`
	Action     map[string]interface{} `json:"action"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// RoutingRuleCreateRequest represents a routing rule create payload.
type RoutingRuleCreateRequest struct {
	Name       string                 `json:"name"`
	Status     string                 `json:"status,omitempty"`
	Priority   int                    `json:"priority,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Action     map[string]interface{} `json:"action,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RoutingRuleUpdateRequest represents a routing rule update payload.
type RoutingRuleUpdateRequest struct {
	Name       string                 `json:"name,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Priority   int                    `json:"priority,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Action     map[string]interface{} `json:"action,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// List returns routing rules for the current org.
func (s *RoutingRulesService) List(ctx context.Context) ([]RoutingRule, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, "/v1/routing-rules", nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[RoutingRule](raw, "rules")
}

// Create creates a routing rule.
func (s *RoutingRulesService) Create(ctx context.Context, req *RoutingRuleCreateRequest, opts ...RequestOption) (*RoutingRule, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/routing-rules", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var rule RoutingRule
	if err := s.client.do(ctx, httpRequest, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// Get fetches a routing rule by ID.
func (s *RoutingRulesService) Get(ctx context.Context, ruleID string) (*RoutingRule, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/routing-rules/%s", ruleID), nil)
	if err != nil {
		return nil, err
	}

	var rule RoutingRule
	if err := s.client.do(ctx, httpRequest, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// Update updates a routing rule.
func (s *RoutingRulesService) Update(ctx context.Context, ruleID string, req *RoutingRuleUpdateRequest, opts ...RequestOption) (*RoutingRule, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/routing-rules/%s", ruleID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var rule RoutingRule
	if err := s.client.do(ctx, httpRequest, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// Delete removes a routing rule.
func (s *RoutingRulesService) Delete(ctx context.Context, ruleID string, opts ...RequestOption) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, fmt.Sprintf("/v1/routing-rules/%s", ruleID), nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	return s.client.do(ctx, httpRequest, nil)
}
