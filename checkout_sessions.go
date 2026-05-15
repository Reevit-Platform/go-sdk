package reevit

import (
	"context"
	"net/http"
	"time"
)

// CheckoutSessionsService handles server-created checkout sessions for browser SDKs.
type CheckoutSessionsService service

// CheckoutSession represents a server-created checkout session.
type CheckoutSession struct {
	ID            string    `json:"id"`
	ClientSecret  string    `json:"client_secret"`
	SessionSecret string    `json:"session_secret"`
	PaymentIntent *Payment  `json:"payment_intent"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// Create creates a checkout session that can be handed to browser SDKs.
//
// API Docs: POST /v1/checkout/sessions
func (s *CheckoutSessionsService) Create(ctx context.Context, req *PaymentIntentRequest, opts ...RequestOption) (*CheckoutSession, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/checkout/sessions", req)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var session CheckoutSession
	if err := s.client.do(ctx, httpRequest, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
