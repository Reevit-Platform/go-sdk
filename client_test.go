package reevit

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// An injected http.Client is the normal way to attach a tracing or proxying
// Transport. Before this guard it also silently dropped the SDK's timeout, so
// a hung PSP round-trip blocked the caller's goroutine forever.
func TestInjectedHTTPClientHonoursSDKTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pay_slow"}`))
	}))
	defer server.Close()

	injected := &http.Client{}
	client := NewClient("pfk_test_key", "org_123",
		WithBaseURL(server.URL),
		WithHTTPClient(injected),
		WithTimeout(10*time.Millisecond),
	)

	_, err := client.Payments.Get(context.Background(), "pay_slow")
	if err == nil {
		t.Fatal("expected the injected client to time out after 10ms")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("err = %v, want a timeout error", err)
	}
	if injected.Timeout != 0 {
		t.Fatalf("the caller's client was mutated: Timeout = %v", injected.Timeout)
	}
}

func TestInjectedHTTPClientInheritsDefaultTimeout(t *testing.T) {
	t.Parallel()

	client := NewClient("pfk_test_key", "org_123", WithHTTPClient(&http.Client{}))
	if client.httpClient.Timeout != defaultTimeout {
		t.Fatalf("Timeout = %v, want the %v default", client.httpClient.Timeout, defaultTimeout)
	}
}

func TestInjectedHTTPClientKeepsItsOwnTimeout(t *testing.T) {
	t.Parallel()

	client := NewClient("pfk_test_key", "org_123", WithHTTPClient(&http.Client{Timeout: 5 * time.Second}))
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %v, want the caller's 5s", client.httpClient.Timeout)
	}
}

func TestWithTimeoutWinsOverInjectedClientInEitherOrder(t *testing.T) {
	t.Parallel()

	injected := &http.Client{Timeout: 5 * time.Second}

	after := NewClient("pfk_test_key", "org_123", WithHTTPClient(injected), WithTimeout(time.Second))
	if after.httpClient.Timeout != time.Second {
		t.Fatalf("Timeout = %v, want 1s", after.httpClient.Timeout)
	}

	before := NewClient("pfk_test_key", "org_123", WithTimeout(time.Second), WithHTTPClient(injected))
	if before.httpClient.Timeout != time.Second {
		t.Fatalf("Timeout = %v, want 1s", before.httpClient.Timeout)
	}
}

func TestDefaultClientKeepsThirtySecondTimeout(t *testing.T) {
	t.Parallel()

	client := NewClient("pfk_test_key", "org_123")
	if client.httpClient.Timeout != 30*time.Second {
		t.Fatalf("Timeout = %v, want 30s", client.httpClient.Timeout)
	}
}

// An id is merchant-controlled input. Before pathf it was interpolated raw, so
// an id containing "/" or "?" rewrote the request path.
func TestPathSegmentsAreEscapedOnTheWire(t *testing.T) {
	t.Parallel()

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pay_1"}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	if _, err := client.Payments.Get(context.Background(), "pay_1/../../v1/admin?x=1"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	want := "/v1/payments/pay_1%2F..%2F..%2Fv1%2Fadmin%3Fx=1"
	if gotPath != want {
		t.Fatalf("path = %q, want %q", gotPath, want)
	}
}

func TestAPIErrorCapturesRequestID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		body   string
		want   string
	}{
		{
			name:   "x-request-id on a structured error",
			header: "X-Request-Id",
			body:   `{"code":"payment_not_found","message":"no such payment"}`,
			want:   "req_abc123",
		},
		{
			name:   "x-reevit-request-id fallback",
			header: "X-Reevit-Request-Id",
			body:   `{"code":"payment_not_found","message":"no such payment"}`,
			want:   "req_abc123",
		},
		{
			name:   "non-JSON error body",
			header: "X-Request-Id",
			body:   `<html>bad gateway</html>`,
			want:   "req_abc123",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(tt.header, "req_abc123")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
			_, err := client.Payments.Get(context.Background(), "pay_missing")

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %v, want *APIError", err)
			}
			if apiErr.RequestID != tt.want {
				t.Fatalf("RequestID = %q, want %q", apiErr.RequestID, tt.want)
			}
			if !strings.Contains(apiErr.Error(), "[request id: req_abc123]") {
				t.Fatalf("Error() = %q, want it to quote the request id", apiErr.Error())
			}
		})
	}
}

func TestAPIErrorOmitsRequestIDWhenAbsent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"invalid_request","message":"bad"}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	_, err := client.Payments.Get(context.Background(), "pay_1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.RequestID != "" {
		t.Fatalf("RequestID = %q, want empty", apiErr.RequestID)
	}
	if strings.Contains(apiErr.Error(), "request id") {
		t.Fatalf("Error() = %q, want no request id clause", apiErr.Error())
	}
}
