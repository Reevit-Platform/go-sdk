package reevit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURLProduction = "https://api.reevit.io"
	defaultBaseURLSandbox    = "https://sandbox-api.reevit.io"
	userAgent                = "@reevit/go"
)

// Client is the Reevit API client.
type Client struct {
	baseURL    string
	apiKey     string
	orgID      string
	httpClient *http.Client

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services
	Payments      *PaymentsService
	Connections   *ConnectionsService
	Subscriptions *SubscriptionsService
	Fraud         *FraudService
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
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient returns a new Reevit API client.
func NewClient(apiKey, orgID string, opts ...Option) *Client {
	c := &Client{
		baseURL: "",
		apiKey:  apiKey,
		orgID:   orgID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	if strings.HasPrefix(c.apiKey, "pk_test_") || strings.HasPrefix(c.apiKey, "pk_sandbox_") {
		c.baseURL = defaultBaseURLSandbox
	} else {
		c.baseURL = defaultBaseURLProduction
	}

	for _, opt := range opts {
		opt(c)
	}

	c.common.client = c
	c.Payments = (*PaymentsService)(&c.common)
	c.Connections = (*ConnectionsService)(&c.common)
	c.Subscriptions = (*SubscriptionsService)(&c.common)
	c.Fraud = (*FraudService)(&c.common)

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
	u := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(path, "/"))

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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Reevit-Client", "@reevit/go")

	return req, nil
}

// do executes an API request.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for API errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// APIError represents a Reevit API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("reevit: request failed with status %d: %s", e.StatusCode, e.Message)
}
