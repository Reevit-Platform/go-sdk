package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// CustomersService handles communication with customer related methods of the Reevit API.
type CustomersService service

// Customer represents a customer in Reevit.
type Customer struct {
	ID         string                 `json:"id"`
	ExternalID string                 `json:"external_id"`
	Email      string                 `json:"email"`
	Phone      string                 `json:"phone"`
	Name       string                 `json:"name"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateCustomerRequest represents a request to create a customer.
type CreateCustomerRequest struct {
	ExternalID string                 `json:"external_id,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Phone      string                 `json:"phone,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCustomerRequest represents a partial customer update.
type UpdateCustomerRequest struct {
	ExternalID string                 `json:"external_id,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Phone      string                 `json:"phone,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CustomerListOptions contains supported filters for customer listing.
type CustomerListOptions struct {
	Limit      int
	Offset     int
	Email      string
	ExternalID string
}

// TopCustomersOptions contains filters for top customer queries.
type TopCustomersOptions struct {
	Limit    int
	Currency string
	Country  string
	From     string
	To       string
}

// PaginationOptions contains basic offset pagination filters.
type PaginationOptions struct {
	Limit  int
	Offset int
}

// List returns customers for the current org.
func (s *CustomersService) List(ctx context.Context, options ...CustomerListOptions) ([]Customer, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "email", options[0].Email)
		setString(values, "external_id", options[0].ExternalID)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/customers", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[Customer](raw, "customers")
}

// Create creates a new customer.
func (s *CustomersService) Create(ctx context.Context, req *CreateCustomerRequest, opts ...RequestOption) (*Customer, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/customers", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var customer Customer
	if err := s.client.do(ctx, httpRequest, &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// Get fetches a customer by ID.
func (s *CustomersService) Get(ctx context.Context, customerID string) (*Customer, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/customers/%s", customerID), nil)
	if err != nil {
		return nil, err
	}

	var customer Customer
	if err := s.client.do(ctx, httpRequest, &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// Update updates a customer by ID.
func (s *CustomersService) Update(ctx context.Context, customerID string, req *UpdateCustomerRequest, opts ...RequestOption) (*Customer, error) {
	httpRequest, err := s.client.newRequest(http.MethodPatch, fmt.Sprintf("/v1/customers/%s", customerID), req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var customer Customer
	if err := s.client.do(ctx, httpRequest, &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// Delete removes a customer.
func (s *CustomersService) Delete(ctx context.Context, customerID string, opts ...RequestOption) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, fmt.Sprintf("/v1/customers/%s", customerID), nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	return s.client.do(ctx, httpRequest, nil)
}

// Lookup fetches a customer by external ID.
func (s *CustomersService) Lookup(ctx context.Context, externalID string) (*Customer, error) {
	values := url.Values{}
	setString(values, "external_id", externalID)

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/customers/lookup", values), nil)
	if err != nil {
		return nil, err
	}

	var customer Customer
	if err := s.client.do(ctx, httpRequest, &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// Top returns top customers by value or volume.
func (s *CustomersService) Top(ctx context.Context, options ...TopCustomersOptions) ([]Customer, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setString(values, "currency", options[0].Currency)
		setString(values, "country", options[0].Country)
		setString(values, "from", options[0].From)
		setString(values, "to", options[0].To)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/customers/top", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[Customer](raw, "customers")
}

// ListPayments returns payment history for a customer.
func (s *CustomersService) ListPayments(ctx context.Context, customerID string, options ...PaginationOptions) ([]PaymentSummary, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath(fmt.Sprintf("/v1/customers/%s/payments", customerID), values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[PaymentSummary](raw, "payments")
}
