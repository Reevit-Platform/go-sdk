package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// InvoicesService handles invoice related methods of the Reevit API.
type InvoicesService service

// Invoice represents an invoice resource.
type Invoice struct {
	ID         string                 `json:"id"`
	CustomerID string                 `json:"customer_id"`
	Status     string                 `json:"status"`
	Amount     int64                  `json:"amount"`
	Currency   string                 `json:"currency"`
	DueDate    *time.Time             `json:"due_date"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// InvoiceListOptions contains supported list filters.
type InvoiceListOptions struct {
	Limit      int
	Offset     int
	Status     string
	CustomerID string
}

// InvoiceUpdateRequest represents a partial invoice update payload.
type InvoiceUpdateRequest struct {
	Status   string                 `json:"status,omitempty"`
	DueDate  string                 `json:"due_date,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// List returns invoices for the current org.
func (s *InvoicesService) List(ctx context.Context, options ...InvoiceListOptions) ([]Invoice, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "status", options[0].Status)
		setString(values, "customer_id", options[0].CustomerID)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/invoices", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[Invoice](raw, "invoices")
}

// Get fetches an invoice by ID.
func (s *InvoicesService) Get(ctx context.Context, invoiceID string) (*Invoice, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/invoices/%s", invoiceID), nil)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	if err := s.client.do(ctx, httpRequest, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Update updates an invoice.
func (s *InvoicesService) Update(ctx context.Context, invoiceID string, req *InvoiceUpdateRequest, opts ...RequestOption) (*Invoice, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/invoices/%s", invoiceID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var invoice Invoice
	if err := s.client.do(ctx, httpRequest, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Cancel cancels an invoice.
func (s *InvoicesService) Cancel(ctx context.Context, invoiceID string, opts ...RequestOption) (*Invoice, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/invoices/%s/cancel", invoiceID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var invoice Invoice
	if err := s.client.do(ctx, httpRequest, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Retry retries invoice collection.
func (s *InvoicesService) Retry(ctx context.Context, invoiceID string, opts ...RequestOption) (*Invoice, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/invoices/%s/retry", invoiceID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var invoice Invoice
	if err := s.client.do(ctx, httpRequest, &invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}
