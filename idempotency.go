package reevit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GenerateIdempotencyKey creates a deterministic idempotency key from input parameters.
// It uses a stable key ordering and a 5-minute time bucket (matching JS SDK behavior).
func GenerateIdempotencyKey(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for i, key := range keys {
		if i > 0 {
			builder.WriteString("|")
		}
		valueBytes, _ := json.Marshal(params[key])
		builder.WriteString(key)
		builder.WriteString(":")
		builder.Write(valueBytes)
	}

	sum := sha256.Sum256([]byte(builder.String()))
	timeBucket := time.Now().Unix() / int64(5*60)

	return fmt.Sprintf("reevit_%d_%x", timeBucket, sum)
}
