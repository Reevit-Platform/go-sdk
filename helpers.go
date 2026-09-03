package reevit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/"
	}
	return "/" + strings.TrimLeft(trimmed, "/")
}

func isPublicPath(path string) bool {
	normalized := normalizePath(path)
	return strings.HasPrefix(normalized, "/v1/pay/")
}

// pathf builds a request path from a format string whose verbs are all %s,
// percent-encoding every segment with url.PathEscape. Interpolating an id
// straight into a path lets a value containing "/", "?" or "#" rewrite the
// request path, so every path with an interpolated identifier must go through
// this helper rather than fmt.Sprintf.
func pathf(format string, segs ...string) string {
	escaped := make([]any, len(segs))
	for i, seg := range segs {
		escaped[i] = url.PathEscape(seg)
	}
	return fmt.Sprintf(format, escaped...)
}

func buildPath(path string, values url.Values) string {
	normalized := normalizePath(path)
	encoded := values.Encode()
	if encoded == "" {
		return normalized
	}
	return normalized + "?" + encoded
}

func setString(values url.Values, key, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		values.Set(key, trimmed)
	}
}

func setInt(values url.Values, key string, value int) {
	if value > 0 {
		values.Set(key, strconv.Itoa(value))
	}
}

func setInt64(values url.Values, key string, value int64) {
	if value > 0 {
		values.Set(key, strconv.FormatInt(value, 10))
	}
}

func setBool(values url.Values, key string, value *bool) {
	if value != nil {
		values.Set(key, strconv.FormatBool(*value))
	}
}

// decodeArrayResponse accepts several list response shapes so the SDK stays
// forward-compatible as the backend migrates to a paginated envelope:
//
//  1. a bare array
//  2. {"<key>": [...]} (legacy flat key)
//  3. {"data": [...]} (new envelope, pagination metadata ignored)
//  4. {"data": {"<key>": [...]}} (double-nested, already used by some endpoints)
//
// The legacy flat key is checked before "data" so that today's responses
// resolve at step 2 and this change is a provable no-op against the current
// server. Each candidate is tried in order and the first one that unmarshals
// cleanly into []T wins; if none do, the caller gets an *APIError with code
// "unexpected_response_shape" -- never an empty slice, which a reconciliation
// sweep would misread as "no records".
func decodeArrayResponse[T any](body []byte, key string) ([]T, error) {
	var direct []T
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}

	if raw, ok := wrapped[key]; ok {
		if err := json.Unmarshal(raw, &direct); err == nil {
			return direct, nil
		}
	}

	if raw, ok := wrapped["data"]; ok {
		if err := json.Unmarshal(raw, &direct); err == nil {
			return direct, nil
		}

		var nested map[string]json.RawMessage
		if err := json.Unmarshal(raw, &nested); err == nil {
			if inner, ok := nested[key]; ok {
				if err := json.Unmarshal(inner, &direct); err == nil {
					return direct, nil
				}
			}
		}
	}

	return nil, &APIError{
		Code:    "unexpected_response_shape",
		Message: fmt.Sprintf("response did not include %q or a usable %q envelope", key, "data"),
		Details: map[string]interface{}{"key": key},
	}
}
