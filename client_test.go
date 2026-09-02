package reevit

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
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
