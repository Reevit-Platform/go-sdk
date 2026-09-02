package reevit

import (
	"reflect"
	"testing"
)

type helperTestItem struct {
	ID string `json:"id"`
}

func TestDecodeArrayResponseAcceptsKnownShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		key  string
		want []helperTestItem
	}{
		{
			name: "bare array",
			body: `[{"id":"a"},{"id":"b"}]`,
			key:  "customers",
			want: []helperTestItem{{ID: "a"}, {ID: "b"}},
		},
		{
			name: "legacy flat key",
			body: `{"customers":[{"id":"a"},{"id":"b"}]}`,
			key:  "customers",
			want: []helperTestItem{{ID: "a"}, {ID: "b"}},
		},
		{
			name: "new envelope with pagination",
			body: `{"data":[{"id":"a"},{"id":"b"}],"pagination":{"total":2}}`,
			key:  "customers",
			want: []helperTestItem{{ID: "a"}, {ID: "b"}},
		},
		{
			name: "double-nested envelope",
			body: `{"data":{"logs":[{"id":"a"},{"id":"b"}]},"pagination":{"total":2}}`,
			key:  "logs",
			want: []helperTestItem{{ID: "a"}, {ID: "b"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := decodeArrayResponse[helperTestItem]([]byte(tt.body), tt.key)
			if err != nil {
				t.Fatalf("decodeArrayResponse: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("decodeArrayResponse = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDecodeArrayResponseErrorsOnTotalMiss(t *testing.T) {
	t.Parallel()

	_, err := decodeArrayResponse[helperTestItem]([]byte(`{"other":[{"id":"a"}]}`), "customers")
	if err == nil {
		t.Fatal("expected an error for a response with no matching key")
	}
}

func TestDecodeArrayResponsePrefersLegacyKeyOverData(t *testing.T) {
	t.Parallel()

	// When both the legacy flat key and "data" are present, the legacy key
	// must win so this change is a no-op against today's server responses.
	body := `{"customers":[{"id":"legacy"}],"data":[{"id":"envelope"}]}`

	got, err := decodeArrayResponse[helperTestItem]([]byte(body), "customers")
	if err != nil {
		t.Fatalf("decodeArrayResponse: %v", err)
	}
	want := []helperTestItem{{ID: "legacy"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decodeArrayResponse = %+v, want %+v", got, want)
	}
}

func TestPathfEscapesEverySegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
		segs   []string
		want   string
	}{
		{
			name:   "plain id",
			format: "/v1/payments/%s",
			segs:   []string{"pay_123"},
			want:   "/v1/payments/pay_123",
		},
		{
			name:   "id with a slash cannot add a path segment",
			format: "/v1/payments/%s",
			segs:   []string{"pay_123/refund"},
			want:   "/v1/payments/pay_123%2Frefund",
		},
		{
			name:   "id with a query or fragment cannot truncate the path",
			format: "/v1/payments/%s/confirm",
			segs:   []string{"pay?a=1#b"},
			want:   "/v1/payments/pay%3Fa=1%23b/confirm",
		},
		{
			name:   "multiple segments",
			format: "/v1/customers/%s/payments/%s",
			segs:   []string{"cus/1", "pay/2"},
			want:   "/v1/customers/cus%2F1/payments/pay%2F2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := pathf(tt.format, tt.segs...); got != tt.want {
				t.Fatalf("pathf = %q, want %q", got, tt.want)
			}
		})
	}
}
