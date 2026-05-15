package reevit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WebhooksService handles communication with webhook related methods of the Reevit API.
type WebhooksService service

// WebhookConfig represents the org-level outbound webhook configuration.
type WebhookConfig struct {
	ID        string                 `json:"id"`
	URL       string                 `json:"url"`
	Status    string                 `json:"status"`
	Events    []string               `json:"events"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// WebhookConfigRequest represents a webhook config upsert payload.
type WebhookConfigRequest struct {
	URL      string                 `json:"url"`
	Events   []string               `json:"events,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookEvent represents a recorded webhook event.
type WebhookEvent struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	AttemptCount int                    `json:"attempt_count"`
	Data         map[string]interface{} `json:"data"`
	CreatedAt    time.Time              `json:"created_at"`
}

// OutboundWebhook represents an outbound webhook delivery.
type OutboundWebhook struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	ResponseCode int       `json:"response_code"`
	AttemptCount int       `json:"attempt_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WebhookEventListOptions contains filters for webhook event listing.
type WebhookEventListOptions struct {
	Limit  int
	Offset int
	Type   string
	Status string
}

// GetConfig fetches the current outbound webhook configuration.
func (s *WebhooksService) GetConfig(ctx context.Context) (*WebhookConfig, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, "/v1/webhooks/config", nil)
	if err != nil {
		return nil, err
	}

	var config WebhookConfig
	if err := s.client.do(ctx, httpRequest, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpsertConfig creates or updates the outbound webhook configuration.
func (s *WebhooksService) UpsertConfig(ctx context.Context, req *WebhookConfigRequest, opts ...RequestOption) (*WebhookConfig, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/webhooks/config", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var config WebhookConfig
	if err := s.client.do(ctx, httpRequest, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DeleteConfig removes the outbound webhook configuration.
func (s *WebhooksService) DeleteConfig(ctx context.Context, opts ...RequestOption) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, "/v1/webhooks/config", nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	return s.client.do(ctx, httpRequest, nil)
}

// SendTest dispatches a test webhook.
func (s *WebhooksService) SendTest(ctx context.Context, opts ...RequestOption) (map[string]interface{}, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/webhooks/test", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var payload map[string]interface{}
	if err := s.client.do(ctx, httpRequest, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// ListEvents returns recorded webhook events.
func (s *WebhooksService) ListEvents(ctx context.Context, options ...WebhookEventListOptions) ([]WebhookEvent, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
		setString(values, "type", options[0].Type)
		setString(values, "status", options[0].Status)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/webhooks/events", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[WebhookEvent](raw, "events")
}

// GetEvent fetches a single webhook event.
func (s *WebhooksService) GetEvent(ctx context.Context, eventID string) (*WebhookEvent, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/webhooks/events/%s", eventID), nil)
	if err != nil {
		return nil, err
	}

	var event WebhookEvent
	if err := s.client.do(ctx, httpRequest, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// ReplayEvent replays a recorded webhook event.
func (s *WebhooksService) ReplayEvent(ctx context.Context, eventID string, opts ...RequestOption) (map[string]interface{}, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, fmt.Sprintf("/v1/webhooks/events/%s/replay", eventID), map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var payload map[string]interface{}
	if err := s.client.do(ctx, httpRequest, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// ListOutbound returns outbound deliveries.
func (s *WebhooksService) ListOutbound(ctx context.Context, options ...PaginationOptions) ([]OutboundWebhook, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}

	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/webhooks/outbound", values), nil)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.doRaw(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	return decodeArrayResponse[OutboundWebhook](raw, "outbound")
}

// GetOutbound fetches a single outbound delivery.
func (s *WebhooksService) GetOutbound(ctx context.Context, outboundID string) (*OutboundWebhook, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, fmt.Sprintf("/v1/webhooks/outbound/%s", outboundID), nil)
	if err != nil {
		return nil, err
	}

	var outbound OutboundWebhook
	if err := s.client.do(ctx, httpRequest, &outbound); err != nil {
		return nil, err
	}

	return &outbound, nil
}
