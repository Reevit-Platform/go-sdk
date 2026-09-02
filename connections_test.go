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

// ListPage is the only list endpoint that used to decode its own envelope, so
// it was the one that would break when the backend ships {"data":[...]}. It
// now goes through decodeArrayResponse and must accept every shape.
func TestConnectionsListPageAcceptsEveryEnvelopeShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		wantIDs        []string
		wantPagination ConnectionPagination
	}{
		{
			name:           "bare array",
			body:           `[{"id":"conn_1"},{"id":"conn_2"}]`,
			wantIDs:        []string{"conn_1", "conn_2"},
			wantPagination: ConnectionPagination{Total: 2, Limit: 2, Offset: 0},
		},
		{
			name:           "legacy flat key",
			body:           `{"connections":[{"id":"conn_1"}],"pagination":{"total":9,"limit":1,"offset":4}}`,
			wantIDs:        []string{"conn_1"},
			wantPagination: ConnectionPagination{Total: 9, Limit: 1, Offset: 4},
		},
		{
			name:           "data envelope",
			body:           `{"data":[{"id":"conn_1"}],"pagination":{"total":9,"limit":1,"offset":4}}`,
			wantIDs:        []string{"conn_1"},
			wantPagination: ConnectionPagination{Total: 9, Limit: 1, Offset: 4},
		},
		{
			name:           "double-nested envelope",
			body:           `{"data":{"connections":[{"id":"conn_1"}],"pagination":{"total":9,"limit":1,"offset":4}}}`,
			wantIDs:        []string{"conn_1"},
			wantPagination: ConnectionPagination{Total: 9, Limit: 1, Offset: 4},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
			page, err := client.Connections.ListPage(context.Background())
			if err != nil {
				t.Fatalf("ListPage: %v", err)
			}

			ids := make([]string, 0, len(page.Connections))
			for _, connection := range page.Connections {
				ids = append(ids, connection.ID)
			}
			if !reflect.DeepEqual(ids, tt.wantIDs) {
				t.Fatalf("IDs = %v, want %v", ids, tt.wantIDs)
			}
			if page.Pagination != tt.wantPagination {
				t.Fatalf("Pagination = %+v, want %+v", page.Pagination, tt.wantPagination)
			}
		})
	}
}

func TestConnectionsListPageRejectsUnknownShape(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"conn_1"}]}`))
	}))
	defer server.Close()

	client := NewClient("pfk_test_key", "org_123", WithBaseURL(server.URL))
	if _, err := client.Connections.ListPage(context.Background()); err == nil {
		t.Fatal("expected an error for an unrecognised list shape")
	}
}
