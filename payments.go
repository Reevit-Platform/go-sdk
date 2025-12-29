package reevit

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// PaymentsService handles communication with the payment related methods of the Reevit API.
type PaymentsService service

// PaymentIntentRequest represents a request to create a payment intent.
type PaymentIntentRequest struct {
	Amount     int64                  `json:"amount"`
	Currency   string                 `json:"currency"`
	Method     string                 `json:"method"`
	Country    string                 `json:"country"`
	CustomerID string                 `json:"customer_id,omitempty"`
	Reference  string                 `json:"reference,omitempty"`
	Policy     *FraudPolicyInput      `json:"policy,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FraudPolicyInput represents the fraud policy configuration for a payment.
type FraudPolicyInput struct {
	Prefer               []string `json:"prefer,omitempty"`
	MaxAmount            int64    `json:"max_amount,omitempty"`
	BlockedBins          []string `json:"blocked_bins,omitempty"`
	AllowedBins          []string `json:"allowed_bins,omitempty"`
	VelocityMaxPerMinute int      `json:"velocity_max_per_minute,omitempty"`
}

// Payment represents a payment object.
type Payment struct {
	ID            string                 `json:"id"`
	ConnectionID  string                 `json:"connection_id"`
	Provider      string                 `json:"provider"`
	ProviderRefID string                 `json:"provider_ref_id"`
	Method        string                 `json:"method"`
	Status        string                 `json:"status"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	FeeAmount     int64                  `json:"fee_amount"`
	FeeCurrency   string                 `json:"fee_currency"`
	NetAmount     int64                  `json:"net_amount"`
	CustomerID    string                 `json:"customer_id"`
	Metadata      map[string]interface{} `json:"metadata"`
	Route         []PaymentRouteAttempt  `json:"route"`
	Reference     string                 `json:"reference"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// PaymentSummary represents a summary of a payment object.
type PaymentSummary struct {
	ID           string                 `json:"id"`
	ConnectionID string                 `json:"connection_id"`
	Provider     string                 `json:"provider"`
	Method       string                 `json:"method"`
	Status       string                 `json:"status"`
	Amount       int64                  `json:"amount"`
	Currency     string                 `json:"currency"`
	FeeAmount    int64                  `json:"fee_amount"`
	FeeCurrency  string                 `json:"fee_currency"`
	NetAmount    int64                  `json:"net_amount"`
	CustomerID   string                 `json:"customer_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	Reference    string                 `json:"reference"`
	CreatedAt    time.Time              `json:"created_at"`
}

// PaymentRouteAttempt represents a routing attempt.
type PaymentRouteAttempt struct {
	ConnectionID string        `json:"connection_id"`
	Provider     string        `json:"provider"`
	Status       string        `json:"status"`
	Error        string        `json:"error"`
	Labels       []string      `json:"labels"`
	RoutingHints *RoutingHints `json:"routing_hints"`
}

// RoutingHints represents routing preferences.
type RoutingHints struct {
	CountryPreference []string          `json:"country_preference"`
	MethodBias        map[string]string `json:"method_bias"`
	FallbackOnly      bool              `json:"fallback_only"`
}

// CreateIntent creates a new payment intent.
//
// API Docs: POST /v1/payments/intents
func (s *PaymentsService) CreateIntent(ctx context.Context, req *PaymentIntentRequest, opts ...RequestOption) (*Payment, error) {
	path := "/v1/payments/intents"

	httpRequest, err := s.client.newRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var payment Payment
	if err := s.client.do(ctx, httpRequest, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// List returns a list of payments.
//
// API Docs: GET /v1/payments
func (s *PaymentsService) List(ctx context.Context, limit, offset int) ([]PaymentSummary, error) {
	path := fmt.Sprintf("/v1/payments?limit=%d&offset=%d", limit, offset)

	httpRequest, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var payments []PaymentSummary
	if err := s.client.do(ctx, httpRequest, &payments); err != nil {
		return nil, err
	}

	return payments, nil
}

// Get retrieves a payment by ID.
//
// API Docs: GET /v1/payments/{id}
func (s *PaymentsService) Get(ctx context.Context, paymentID string) (*Payment, error) {
	path := fmt.Sprintf("/v1/payments/%s", paymentID)

	httpRequest, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var payment Payment
	if err := s.client.do(ctx, httpRequest, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// Confirm confirms a payment after PSP callback.
//
// API Docs: POST /v1/payments/{id}/confirm
func (s *PaymentsService) Confirm(ctx context.Context, paymentID string) (*Payment, error) {
	path := fmt.Sprintf("/v1/payments/%s/confirm", paymentID)

	httpRequest, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var payment Payment
	if err := s.client.do(ctx, httpRequest, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// Cancel cancels a payment intent.
//
// API Docs: POST /v1/payments/{id}/cancel
func (s *PaymentsService) Cancel(ctx context.Context, paymentID string) (*Payment, error) {
	path := fmt.Sprintf("/v1/payments/%s/cancel", paymentID)

	httpRequest, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var payment Payment
	if err := s.client.do(ctx, httpRequest, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}
