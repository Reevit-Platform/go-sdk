package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// PaymentLinksService handles communication with payment link related methods of the Reevit API.
type PaymentLinksService service

// PaymentLink represents a hosted payment link.
type PaymentLink struct {
	ID          string                 `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Status      string                 `json:"status"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreatePaymentLinkRequest represents a request to create a payment link.
type CreatePaymentLinkRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Reference   string                 `json:"reference,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePaymentLinkRequest represents a partial payment link update.
type UpdatePaymentLinkRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentLinkListOptions contains supported list filters.
type PaymentLinkListOptions struct {
	Limit  int
	Offset int
	Status string
}

// PaymentLinkStats contains aggregate link performance data.
type PaymentLinkStats struct {
	Visits         int64   `json:"visits"`
	PaymentsCount  int64   `json:"payments_count"`
	TotalAmount    int64   `json:"total_amount"`
	ConversionRate float64 `json:"conversion_rate"`
}

// List returns payment links for the current org.
func (s *PaymentLinksService) List(ctx context.Context, options ...PaymentLinkListOptions) ([]PaymentLink, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "status", options[0].Status)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/payment-links", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[PaymentLink](raw, "payment_links")
}

// Create creates a payment link.
func (s *PaymentLinksService) Create(ctx context.Context, req *CreatePaymentLinkRequest, opts ...RequestOption) (*PaymentLink, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/payment-links", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var link PaymentLink
	if err := s.client.do(ctx, httpRequest, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// Get fetches a payment link by ID.
func (s *PaymentLinksService) Get(ctx context.Context, paymentLinkID string) (*PaymentLink, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/payment-links/%s", paymentLinkID), nil)
	if err != nil {
		return nil, err
	}

	var link PaymentLink
	if err := s.client.do(ctx, httpRequest, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// Update updates a payment link.
func (s *PaymentLinksService) Update(ctx context.Context, paymentLinkID string, req *UpdatePaymentLinkRequest, opts ...RequestOption) (*PaymentLink, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/payment-links/%s", paymentLinkID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var link PaymentLink
	if err := s.client.do(ctx, httpRequest, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// Delete removes a payment link.
func (s *PaymentLinksService) Delete(ctx context.Context, paymentLinkID string, opts ...RequestOption) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, fmt.Sprintf("/v1/payment-links/%s", paymentLinkID), nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	return s.client.do(ctx, httpRequest, nil)
}

// GetStats returns aggregate stats for a payment link.
func (s *PaymentLinksService) GetStats(ctx context.Context, paymentLinkID string) (*PaymentLinkStats, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/payment-links/%s/stats", paymentLinkID), nil)
	if err != nil {
		return nil, err
	}

	var stats PaymentLinkStats
	if err := s.client.do(ctx, httpRequest, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// ListPayments lists payments created from a payment link.
func (s *PaymentLinksService) ListPayments(ctx context.Context, paymentLinkID string, options ...PaginationOptions) ([]PaymentSummary, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath(fmt.Sprintf("/v1/payment-links/%s/payments", paymentLinkID), values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[PaymentSummary](raw, "payments")
}

// GetByCode resolves a public payment link by code.
func (s *PaymentLinksService) GetByCode(ctx context.Context, code string) (*PaymentLink, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/pay/%s", code), nil)
	if err != nil {
		return nil, err
	}

	var link PaymentLink
	if err := s.client.do(ctx, httpRequest, &link); err != nil {
		return nil, err
	}

	return &link, nil
}
