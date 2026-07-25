package reevit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestConnectionTestAcceptsBackendOKResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/connections/test" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	valid, err := client.Connections.Test(
		context.Background(),
		&ConnectionRequest{Provider: "paystack", Mode: "sandbox"},
	)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if !valid {
		t.Fatal("expected backend ok response to report valid credentials")
	}
}

func TestConnectionsListAllFollowsPaginationAndFilters(t *testing.T) {
	t.Parallel()

	var offsets []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for key, want := range map[string]string{
			"provider": "paystack",
			"mode":     "live",
			"status":   "active",
			"label":    "primary",
			"limit":    "200",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		offset, _ := strconv.Atoi(query.Get("offset"))
		offsets = append(offsets, offset)
		ids := []string{"conn_1", "conn_2"}
		if offset > 0 {
			ids = []string{"conn_3"}
		}
		connections := make([]map[string]interface{}, 0, len(ids))
		for _, id := range ids {
			connections = append(connections, map[string]interface{}{
				"id": id, "provider": "paystack", "mode": "live", "status": "active",
				"name_match_status": "match",
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"connections": connections,
			"pagination":  map[string]interface{}{"total": 3, "limit": 200, "offset": offset},
		})
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	connections, err := client.Connections.ListAll(context.Background(), ConnectionListOptions{
		Provider: "paystack",
		Mode:     "live",
		Status:   "active",
		Label:    "primary",
	})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	ids := make([]string, 0, len(connections))
	for _, connection := range connections {
		ids = append(ids, connection.ID)
		if connection.NameMatchStatus != "match" {
			t.Errorf("NameMatchStatus = %q", connection.NameMatchStatus)
		}
	}
	if !reflect.DeepEqual(ids, []string{"conn_1", "conn_2", "conn_3"}) {
		t.Fatalf("IDs = %v", ids)
	}
	if !reflect.DeepEqual(offsets, []int{0, 2}) {
		t.Fatalf("offsets = %v", offsets)
	}
}

func TestConnectionsListLabels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/connections/labels" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"label":"primary","total":2}]`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	labels, err := client.Connections.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	if len(labels) != 1 || labels[0].Label != "primary" || labels[0].Total != 2 {
		t.Fatalf("labels = %+v", labels)
	}
}
