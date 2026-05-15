package reevit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

// PaymentIntentUpdateRequest represents a partial update to a payment intent.
type PaymentIntentUpdateRequest struct {
	Amount   *int64                 `json:"amount,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	ClientSecret  string                 `json:"client_secret"`
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

// RefundRequest represents a payment refund request.
type RefundRequest struct {
	Amount int64  `json:"amount,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Refund represents a refund record returned by the API.
type Refund struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Status    string    `json:"status"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaymentStatsOptions contains filters for payment stats queries.
type PaymentStatsOptions struct {
	From     string
	To       string
	Country  string
	Method   string
	Provider string
	Status   string
	Interval string
}

// PaymentStats captures aggregated payment metrics.
type PaymentStats struct {
	Count           int64                  `json:"count"`
	SuccessCount    int64                  `json:"success_count"`
	FailedCount     int64                  `json:"failed_count"`
	PendingCount    int64                  `json:"pending_count"`
	TotalAmount     int64                  `json:"total_amount"`
	FeeAmount       int64                  `json:"fee_amount"`
	NetAmount       int64                  `json:"net_amount"`
	Currency        string                 `json:"currency"`
	SuccessRate     float64                `json:"success_rate"`
	AverageLatency  float64                `json:"average_latency_ms"`
	AdditionalStats map[string]interface{} `json:"-"`
}

// CreateIntent creates a new payment intent.
//
// API Docs: POST /v1/payments/intents
func (s *PaymentsService) CreateIntent(ctx context.Context, req *PaymentIntentRequest, opts ...RequestOption) (*Payment, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/payments/intents", req)
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
	values := url.Values{}
	setInt(values, "limit", limit)
	setInt(values, "offset", offset)

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/payments", values), nil)
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
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", paymentID), nil)
	if err != nil {
		return nil, err
	}

	var payment Payment
	if err := s.client.do(ctx, httpRequest, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// UpdateIntent updates a payment intent.
//
// API Docs: PATCH /v1/payments/intents/{id}
func (s *PaymentsService) UpdateIntent(ctx context.Context, paymentID string, req *PaymentIntentUpdateRequest, opts ...RequestOption) (*Payment, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/payments/intents/%s", paymentID), req)
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

// Confirm confirms a payment after PSP callback.
//
// API Docs: POST /v1/payments/{id}/confirm
func (s *PaymentsService) Confirm(ctx context.Context, paymentID string, opts ...RequestOption) (*Payment, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/payments/%s/confirm", paymentID), map[string]interface{}{})
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

// ConfirmIntent confirms a payment intent using its client secret.
//
// API Docs: POST /v1/payments/{id}/confirm-intent
func (s *PaymentsService) ConfirmIntent(ctx context.Context, paymentID, clientSecret string, opts ...RequestOption) (*Payment, error) {
	values := url.Values{}
	setString(values, "client_secret", clientSecret)

	httpRequest, err := s.client.newRequest(
		http.MethodPost,
		buildPath(fmt.Sprintf("/v1/payments/%s/confirm-intent", paymentID), values),
		map[string]interface{}{},
	)
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

// Cancel cancels a payment.
//
// API Docs: POST /v1/payments/{id}/cancel
func (s *PaymentsService) Cancel(ctx context.Context, paymentID string, opts ...RequestOption) (*Payment, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/payments/%s/cancel", paymentID), map[string]interface{}{})
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

// Retry retries a payment with the orchestration layer.
//
// API Docs: POST /v1/payments/{id}/retry
func (s *PaymentsService) Retry(ctx context.Context, paymentID string, opts ...RequestOption) (*Payment, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/payments/%s/retry", paymentID), map[string]interface{}{})
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

// Refund creates a refund for a payment.
//
// API Docs: POST /v1/payments/{id}/refund
func (s *PaymentsService) Refund(ctx context.Context, paymentID string, req *RefundRequest, opts ...RequestOption) (*Refund, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/payments/%s/refund", paymentID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var refund Refund
	if err := s.client.do(ctx, httpRequest, &refund); err != nil {
		return nil, err
	}

	return &refund, nil
}

// GetStats returns aggregated payment stats.
//
// API Docs: GET /v1/payments/stats
func (s *PaymentsService) GetStats(ctx context.Context, options *PaymentStatsOptions) (*PaymentStats, error) {
	values := url.Values{}
	if options != nil {
		setString(values, "from", options.From)
		setString(values, "to", options.To)
		setString(values, "country", options.Country)
		setString(values, "method", options.Method)
		setString(values, "provider", options.Provider)
		setString(values, "status", options.Status)
		setString(values, "interval", options.Interval)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/payments/stats", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	var stats PaymentStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &stats.AdditionalStats); err != nil {
		stats.AdditionalStats = map[string]interface{}{}
	}

	return &stats, nil
}
