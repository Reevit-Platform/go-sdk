package reevit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.reevit.io"
	userAgent      = "@reevit/go v0.10.1"

	// defaultTimeout matches the Python/PHP SDK defaults; payment
	// confirmation calls block on PSP round-trips that routinely exceed 10s.
	defaultTimeout = 30 * time.Second
)

// Client is the Reevit API client.
type Client struct {
	baseURL    string
	apiKey     string
	orgID      string
	httpClient *http.Client

	// timeout is the per-request timeout applied to httpClient. It is kept
	// on the Client so that a caller-supplied http.Client (WithHTTPClient)
	// still gets a bound, instead of silently becoming unbounded.
	timeout         time.Duration
	timeoutExplicit bool

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services
	Payments         *PaymentsService
	Connections      *ConnectionsService
	Subscriptions    *SubscriptionsService
	Fraud            *FraudService
	Customers        *CustomersService
	PaymentLinks     *PaymentLinksService
	CheckoutSessions *CheckoutSessionsService
	Webhooks         *WebhooksService
	RoutingRules     *RoutingRulesService
	Invoices         *InvoicesService
	Payouts          *PayoutsService
}

type service struct {
	client *Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets the base URL for the API.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient sets the HTTP client used for requests.
//
// The client is shallow-copied so the caller's value is never mutated. If its
// Timeout is zero -- the normal case when a client is injected only to attach
// a custom Transport for tracing or proxying -- the SDK's timeout (30s, or
// whatever WithTimeout sets) is applied to the copy. Without that, a hung PSP
// round-trip would block the calling goroutine forever and leak connections.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient == nil {
			return
		}
		clone := *httpClient
		c.httpClient = &clone
	}
}

// WithTimeout overrides the per-request timeout (30s by default).
//
// An explicit WithTimeout wins over the Timeout of a client passed to
// WithHTTPClient regardless of option order. Non-positive durations are
// ignored; use context.WithCancel if you need an unbounded request.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout <= 0 {
			return
		}
		c.timeout = timeout
		c.timeoutExplicit = true
	}
}

// NewClient returns a new Reevit API client.
func NewClient(apiKey, orgID string, opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		orgID:      orgID,
		httpClient: &http.Client{},
		timeout:    defaultTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Resolve the effective timeout after every option has run so that
	// WithTimeout and WithHTTPClient compose in either order.
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	if c.timeoutExplicit || c.httpClient.Timeout == 0 {
		c.httpClient.Timeout = c.timeout
	}

	c.common.client = c
	c.Payments = (*PaymentsService)(&c.common)
	c.Connections = (*ConnectionsService)(&c.common)
	c.Subscriptions = (*SubscriptionsService)(&c.common)
	c.Fraud = (*FraudService)(&c.common)
	c.Customers = (*CustomersService)(&c.common)
	c.PaymentLinks = (*PaymentLinksService)(&c.common)
	c.CheckoutSessions = (*CheckoutSessionsService)(&c.common)
	c.Webhooks = (*WebhooksService)(&c.common)
	c.RoutingRules = (*RoutingRulesService)(&c.common)
	c.Invoices = (*InvoicesService)(&c.common)
	c.Payouts = (*PayoutsService)(&c.common)

	return c
}

// RequestOption is a functional option for configuring API requests.
type RequestOption func(*http.Request)

// WithIdempotencyKey sets the idempotency key for the request.
func WithIdempotencyKey(key string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("Idempotency-Key", key)
	}
}

// newRequest creates an API request.
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	normalizedPath := normalizePath(path)
	if !isPublicPath(normalizedPath) && strings.TrimSpace(c.orgID) == "" {
		return nil, errors.New("reevit: orgID is required for authenticated requests")
	}
	u := fmt.Sprintf("%s%s", strings.TrimRight(c.baseURL, "/"), normalizedPath)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u, buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Reevit-Client", "@reevit/go")
	req.Header.Set("X-Reevit-Client-Version", "0.10.1")
	if strings.TrimSpace(c.apiKey) != "" {
		req.Header.Set("X-Reevit-Key", c.apiKey)
	}
	if !isPublicPath(normalizedPath) && strings.TrimSpace(c.orgID) != "" {
		req.Header.Set("X-Org-Id", c.orgID)
	}

	return req, nil
}

// do executes an API request.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	body, err := c.doRaw(ctx, req)
	if err != nil {
		return err
	}
	if v == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}

func (c *Client) doRaw(ctx context.Context, req *http.Request) ([]byte, error) {
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		payload := struct {
			Code    string                 `json:"code"`
			Message string                 `json:"message"`
			Details map[string]interface{} `json:"details"`
		}{}
		message := strings.TrimSpace(string(bodyBytes))
		if err := json.Unmarshal(bodyBytes, &payload); err == nil {
			if payload.Message != "" {
				message = payload.Message
			}
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Code:       payload.Code,
				Message:    message,
				Details:    payload.Details,
			}
		}
		if message == "" {
			message = resp.Status
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    message,
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	return bodyBytes, nil
}

// APIError represents a Reevit API error.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("reevit: request failed with status %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("reevit: request failed with status %d: %s", e.StatusCode, e.Message)
}
