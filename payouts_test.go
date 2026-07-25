package reevit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayoutCreateRequiresIdempotencyKey(t *testing.T) {
	t.Parallel()

	client := NewClient("pfk_test_key", "org_123")
	_, err := client.Payouts.Create(context.Background(), &CreatePayoutRequest{})
	if err == nil || err.Error() != "reevit: idempotency key is required for payout creation" {
		t.Fatalf("expected idempotency error, got %v", err)
	}
}

func TestPayoutCreateSendsHeadersAndPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/payouts" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "order-123" {
			t.Errorf("Idempotency-Key = %q", got)
		}
		if got := r.Header.Get("X-Org-Id"); got != "org_123" {
			t.Errorf("X-Org-Id = %q", got)
		}
		var body CreatePayoutRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Amount != 2500 || body.Currency != "GHS" {
			t.Errorf("unexpected body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"po_123","amount":2500,"currency":"GHS","status":"pending","created_at":"2026-07-25T00:00:00Z","updated_at":"2026-07-25T00:00:00Z"}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	payout, err := client.Payouts.Create(
		context.Background(),
		&CreatePayoutRequest{Amount: 2500, Currency: "GHS"},
		WithIdempotencyKey("order-123"),
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if payout.ID != "po_123" {
		t.Fatalf("ID = %q", payout.ID)
	}
}
