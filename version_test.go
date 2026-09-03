package reevit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The SDK version reached the wire from two independent string literals: the
// userAgent const and an inline argument to the X-Reevit-Client-Version
// header. Bumping one and forgetting the other shipped a client that reported
// two different versions of itself in the same request. Both now derive from
// Version; this pins that they agree with it and with each other.
func TestBothVersionHeadersReportTheSameVersion(t *testing.T) {
	t.Parallel()

	var gotUA, gotVersion string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotVersion = r.Header.Get("X-Reevit-Client-Version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pay_1"}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	if _, err := client.Payments.Get(context.Background(), "pay_1"); err != nil {
		t.Fatalf("get: %v", err)
	}

	if gotVersion != Version {
		t.Errorf("X-Reevit-Client-Version = %q, want %q", gotVersion, Version)
	}
	if !strings.HasSuffix(gotUA, "v"+Version) {
		t.Errorf("User-Agent = %q, want it to end with %q", gotUA, "v"+Version)
	}
}
