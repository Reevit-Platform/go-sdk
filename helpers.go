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

func decodeArrayResponse[T any](body []byte, key string) ([]T, error) {
	var direct []T
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}

	raw, ok := wrapped[key]
	if !ok {
		return nil, fmt.Errorf("reevit: response did not include %q", key)
	}

	if err := json.Unmarshal(raw, &direct); err != nil {
		return nil, err
	}

	return direct, nil
}
