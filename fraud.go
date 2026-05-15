package reevit

import (
	"context"
	"net/http"
)

// FraudService handles communication with the fraud policy related methods of the Reevit API.
type FraudService service

// FraudPolicy represents the fraud policy configuration.
type FraudPolicy struct {
	Prefer               []string `json:"prefer"`
	MaxAmount            int64    `json:"max_amount"`
	BlockedBins          []string `json:"blocked_bins"`
	AllowedBins          []string `json:"allowed_bins"`
	VelocityMaxPerMinute int      `json:"velocity_max_per_minute"`
}

// Get retrieves the current fraud policy.
//
// API Docs: GET /v1/policies/fraud
func (s *FraudService) Get(ctx context.Context) (*FraudPolicy, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, "/v1/policies/fraud", nil)
	if err != nil {
		return nil, err
	}

	var policy FraudPolicy
	if err := s.client.do(ctx, httpRequest, &policy); err != nil {
		return nil, err
	}

	return &policy, nil
}

// Update updates the fraud policy.
//
// API Docs: POST /v1/policies/fraud
func (s *FraudService) Update(ctx context.Context, policy *FraudPolicy, opts ...RequestOption) (*FraudPolicy, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/policies/fraud", policy)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(httpRequest)
	}

	var updatedPolicy FraudPolicy
	if err := s.client.do(ctx, httpRequest, &updatedPolicy); err != nil {
		return nil, err
	}

	return &updatedPolicy, nil
}
