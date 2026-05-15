package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SubscriptionsService handles communication with the subscription related methods of the Reevit API.
type SubscriptionsService service

// SubscriptionRequest represents a request to create a subscription.
type SubscriptionRequest struct {
	CustomerID string                 `json:"customer_id"`
	PlanID     string                 `json:"plan_id"`
	Amount     int64                  `json:"amount"`
	Currency   string                 `json:"currency"`
	Method     string                 `json:"method"`
	Interval   string                 `json:"interval"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SubscriptionUpdateRequest represents a partial update to a subscription.
type SubscriptionUpdateRequest struct {
	PlanID   string                 `json:"plan_id,omitempty"`
	Method   string                 `json:"method,omitempty"`
	Interval string                 `json:"interval,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SubscriptionListOptions contains list filters for subscriptions.
type SubscriptionListOptions struct {
	Limit      int
	Offset     int
	Status     string
	CustomerID string
	PlanID     string
}

// Subscription represents a subscription object.
type Subscription struct {
	ID            string                 `json:"id"`
	OrgID         string                 `json:"org_id"`
	CustomerID    string                 `json:"customer_id"`
	PlanID        string                 `json:"plan_id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Method        string                 `json:"method"`
	Interval      string                 `json:"interval"`
	Status        string                 `json:"status"`
	NextRenewalAt time.Time              `json:"next_renewal_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Create creates a new subscription.
//
// API Docs: POST /v1/subscriptions
func (s *SubscriptionsService) Create(ctx context.Context, req *SubscriptionRequest, opts ...RequestOption) (*Subscription, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/subscriptions", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var subscription Subscription
	if err := s.client.do(ctx, httpRequest, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}

// List returns a list of subscriptions.
//
// API Docs: GET /v1/subscriptions
func (s *SubscriptionsService) List(ctx context.Context, options ...SubscriptionListOptions) ([]Subscription, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "status", options[0].Status)
		setString(values, "customer_id", options[0].CustomerID)
		setString(values, "plan_id", options[0].PlanID)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/subscriptions", values), nil)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	if err := s.client.do(ctx, httpRequest, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// Get retrieves a subscription by ID.
//
// API Docs: GET /v1/subscriptions/{id}
func (s *SubscriptionsService) Get(ctx context.Context, subscriptionID string) (*Subscription, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/subscriptions/%s", subscriptionID), nil)
	if err != nil {
		return nil, err
	}

	var subscription Subscription
	if err := s.client.do(ctx, httpRequest, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}

// Update updates a subscription.
//
// API Docs: PATCH /v1/subscriptions/{id}
func (s *SubscriptionsService) Update(ctx context.Context, subscriptionID string, req *SubscriptionUpdateRequest, opts ...RequestOption) (*Subscription, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/subscriptions/%s", subscriptionID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var subscription Subscription
	if err := s.client.do(ctx, httpRequest, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}

// Cancel cancels a subscription.
//
// API Docs: POST /v1/subscriptions/{id}/cancel
func (s *SubscriptionsService) Cancel(ctx context.Context, subscriptionID string, opts ...RequestOption) (*Subscription, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/subscriptions/%s/cancel", subscriptionID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var subscription Subscription
	if err := s.client.do(ctx, httpRequest, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}

// Resume resumes a canceled subscription.
//
// API Docs: POST /v1/subscriptions/{id}/resume
func (s *SubscriptionsService) Resume(ctx context.Context, subscriptionID string, opts ...RequestOption) (*Subscription, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/subscriptions/%s/resume", subscriptionID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var subscription Subscription
	if err := s.client.do(ctx, httpRequest, &subscription); err != nil {
		return nil, err
	}

	return &subscription, nil
}
