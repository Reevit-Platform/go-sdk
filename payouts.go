package reevit

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PayoutsService handles disbursements and saved beneficiaries.
type PayoutsService service

type Beneficiary struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number,omitempty"`
	BankCode      string `json:"bank_code,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Provider      string `json:"provider,omitempty"`
}

type Payout struct {
	ID            string                 `json:"id"`
	OrgID         string                 `json:"org_id"`
	Mode          string                 `json:"mode"`
	ConnectionID  string                 `json:"connection_id"`
	Provider      string                 `json:"provider"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	Beneficiary   Beneficiary            `json:"beneficiary"`
	Narration     string                 `json:"narration,omitempty"`
	Reference     string                 `json:"reference"`
	ProviderRef   string                 `json:"provider_ref,omitempty"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type CreatePayoutRequest struct {
	ConnectionID  string                 `json:"connection_id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Beneficiary   *Beneficiary           `json:"beneficiary,omitempty"`
	BeneficiaryID string                 `json:"beneficiary_id,omitempty"`
	Narration     string                 `json:"narration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type PayoutListOptions struct {
	Status string
	Limit  int
	Offset int
}

type PayoutListResponse struct {
	Payouts []Payout `json:"payouts"`
	Total   int64    `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
}

type BulkPayoutItem struct {
	Amount        int64                  `json:"amount"`
	Beneficiary   *Beneficiary           `json:"beneficiary,omitempty"`
	BeneficiaryID string                 `json:"beneficiary_id,omitempty"`
	Narration     string                 `json:"narration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type BulkPayoutRequest struct {
	ConnectionID string           `json:"connection_id"`
	Currency     string           `json:"currency"`
	Payouts      []BulkPayoutItem `json:"payouts"`
}

type BulkPayoutResult struct {
	Index  int     `json:"index"`
	Payout *Payout `json:"payout,omitempty"`
	Error  string  `json:"error,omitempty"`
}

type BulkPayoutResponse struct {
	Results   []BulkPayoutResult `json:"results"`
	Total     int                `json:"total"`
	Succeeded int                `json:"succeeded"`
	Failed    int                `json:"failed"`
}

type PayoutBalance struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
}

type AccountResolution struct {
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number,omitempty"`
}

type SavedBeneficiary struct {
	ID          string      `json:"id"`
	OrgID       string      `json:"org_id"`
	Mode        string      `json:"mode"`
	Beneficiary Beneficiary `json:"beneficiary"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type BeneficiaryListResponse struct {
	Beneficiaries []SavedBeneficiary `json:"beneficiaries"`
	Total         int64              `json:"total"`
	Limit         int                `json:"limit"`
	Offset        int                `json:"offset"`
}

func (s *PayoutsService) Create(ctx context.Context, request *CreatePayoutRequest, opts ...RequestOption) (*Payout, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/payouts", request)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(httpRequest)
	}
	if strings.TrimSpace(httpRequest.Header.Get("Idempotency-Key")) == "" {
		return nil, errors.New("reevit: idempotency key is required for payout creation")
	}

	var payout Payout
	if err := s.client.do(ctx, httpRequest, &payout); err != nil {
		return nil, err
	}
	return &payout, nil
}

func (s *PayoutsService) List(ctx context.Context, options ...PayoutListOptions) (*PayoutListResponse, error) {
	values := url.Values{}
	if len(options) > 0 {
		setString(values, "status", options[0].Status)
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}
	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/payouts", values), nil)
	if err != nil {
		return nil, err
	}
	var response PayoutListResponse
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) Get(ctx context.Context, payoutID string) (*Payout, error) {
	return s.payoutAction(ctx, http.MethodGet, pathf("/v1/payouts/%s", payoutID))
}

func (s *PayoutsService) Confirm(ctx context.Context, payoutID string) (*Payout, error) {
	return s.payoutAction(ctx, http.MethodPost, pathf("/v1/payouts/%s/confirm", payoutID))
}

func (s *PayoutsService) Cancel(ctx context.Context, payoutID string) (*Payout, error) {
	return s.payoutAction(ctx, http.MethodPost, pathf("/v1/payouts/%s/cancel", payoutID))
}

func (s *PayoutsService) payoutAction(ctx context.Context, method, path string) (*Payout, error) {
	var body interface{}
	if method == http.MethodPost {
		body = map[string]interface{}{}
	}
	httpRequest, err := s.client.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	var payout Payout
	if err := s.client.do(ctx, httpRequest, &payout); err != nil {
		return nil, err
	}
	return &payout, nil
}

func (s *PayoutsService) CreateBulk(ctx context.Context, request *BulkPayoutRequest, opts ...RequestOption) (*BulkPayoutResponse, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/payouts/bulk", request)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(httpRequest)
	}
	if strings.TrimSpace(httpRequest.Header.Get("Idempotency-Key")) == "" {
		return nil, errors.New("reevit: idempotency key is required for payout creation")
	}
	var response BulkPayoutResponse
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) Balance(ctx context.Context, connectionID string) ([]PayoutBalance, error) {
	values := url.Values{"connection_id": []string{connectionID}}
	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/payouts/balance", values), nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Balances []PayoutBalance `json:"balances"`
	}
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return response.Balances, nil
}

func (s *PayoutsService) ResolveAccount(ctx context.Context, connectionID string, beneficiary Beneficiary) (*AccountResolution, error) {
	body := map[string]interface{}{"connection_id": connectionID, "beneficiary": beneficiary}
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/payouts/resolve-account", body)
	if err != nil {
		return nil, err
	}
	var response AccountResolution
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) CreateBeneficiary(ctx context.Context, beneficiary Beneficiary) (*SavedBeneficiary, error) {
	httpRequest, err := s.client.newRequest(http.MethodPost, "/v1/beneficiaries", map[string]interface{}{"beneficiary": beneficiary})
	if err != nil {
		return nil, err
	}
	var response SavedBeneficiary
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) ListBeneficiaries(ctx context.Context, options ...PaginationOptions) (*BeneficiaryListResponse, error) {
	values := url.Values{}
	if len(options) > 0 {
		setInt(values, "limit", options[0].Limit)
		setInt(values, "offset", options[0].Offset)
	}
	httpRequest, err := s.client.newRequest(http.MethodGet, buildPath("/v1/beneficiaries", values), nil)
	if err != nil {
		return nil, err
	}
	var response BeneficiaryListResponse
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) GetBeneficiary(ctx context.Context, beneficiaryID string) (*SavedBeneficiary, error) {
	httpRequest, err := s.client.newRequest(http.MethodGet, pathf("/v1/beneficiaries/%s", beneficiaryID), nil)
	if err != nil {
		return nil, err
	}
	var response SavedBeneficiary
	if err := s.client.do(ctx, httpRequest, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *PayoutsService) DeleteBeneficiary(ctx context.Context, beneficiaryID string) error {
	httpRequest, err := s.client.newRequest(http.MethodDelete, pathf("/v1/beneficiaries/%s", beneficiaryID), nil)
	if err != nil {
		return err
	}
	return s.client.do(ctx, httpRequest, nil)
}
