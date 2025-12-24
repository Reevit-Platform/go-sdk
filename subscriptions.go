package reevit

import (
	"context"
	"net/http"
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
	Interval   string                 `json:"interval"` // monthly, yearly
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
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
	Status        string                 `json:"status"` // active, paused, canceled
	NextRenewalAt time.Time              `json:"next_renewal_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Create creates a new subscription.
//
// API Docs: POST /v1/subscriptions
func (s *SubscriptionsService) Create(ctx context.Context, req *SubscriptionRequest) (*Subscription, error) {
	path := "v1/subscriptions"

	httpRequest, err := s.client.newRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
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
func (s *SubscriptionsService) List(ctx context.Context) ([]Subscription, error) {
	path := "v1/subscriptions"

	httpRequest, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	if err := s.client.do(ctx, httpRequest, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}
