package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SignaturePrefix is the algorithm prefix Reevit sends on the
// X-Reevit-Signature header.
const SignaturePrefix = "sha256="

// DefaultTolerance is the replay window used when VerifyWithTolerance is given
// a non-positive tolerance. It matches the other Reevit SDKs.
const DefaultTolerance = 5 * time.Minute

// SignatureHeader is the header Reevit sends the signature on.
const SignatureHeader = "X-Reevit-Signature"

var (
	// ErrInvalidSignature is returned when the signature does not match the body.
	ErrInvalidSignature = errors.New("reevit: webhook signature is invalid")

	// ErrMissingTimestamp is returned when the signed body carries no
	// signature_timestamp to check against the tolerance.
	ErrMissingTimestamp = errors.New("reevit: webhook payload has no signature_timestamp")

	// ErrTimestampOutsideTolerance is returned when the signature is valid but
	// the delivery is too old (or too far in the future) to accept.
	ErrTimestampOutsideTolerance = errors.New("reevit: webhook signature_timestamp is outside the tolerance")
)

// Sign returns the X-Reevit-Signature header value for a raw webhook body:
// "sha256=" followed by the hex HMAC-SHA256 of the body. Mostly useful in
// tests; the same value is what Verify recomputes.
func Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return SignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

// Verify reports whether signature is a valid X-Reevit-Signature for payload.
//
// payload must be the exact bytes received. Do not unmarshal and re-marshal the
// body first: key order and whitespace must match what Reevit signed. The
// comparison uses hmac.Equal, so it does not leak timing information -- writing
// this check by hand with == is the classic mistake, and it is invisible in
// review.
//
//	if !webhooks.Verify(body, r.Header.Get(webhooks.SignatureHeader), secret) {
//	    http.Error(w, "invalid signature", http.StatusUnauthorized)
//	    return
//	}
func Verify(payload []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}
	return hmac.Equal([]byte(Sign(payload, secret)), []byte(signature))
}

// VerifyWithTolerance verifies the signature and then rejects a delivery whose
// signed timestamp is not within tolerance of now, which stops a captured
// delivery from being replayed later.
//
// signedAt is the signature_timestamp carried in the signed body. Pass the zero
// time to have it read out of payload (RFC 3339). A non-positive tolerance
// falls back to DefaultTolerance. Returns nil when the delivery is acceptable.
func VerifyWithTolerance(payload []byte, signature, secret string, signedAt time.Time, tolerance time.Duration) error {
	if !Verify(payload, signature, secret) {
		return ErrInvalidSignature
	}

	if signedAt.IsZero() {
		parsed, err := signedTimestamp(payload)
		if err != nil {
			return err
		}
		signedAt = parsed
	}

	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}

	skew := time.Since(signedAt)
	if skew < 0 {
		skew = -skew
	}
	if skew > tolerance {
		return fmt.Errorf("%w: signed %s ago, tolerance %s", ErrTimestampOutsideTolerance, skew.Round(time.Second), tolerance)
	}

	return nil
}

// signedTimestamp reads the RFC 3339 signature_timestamp out of a signed body.
func signedTimestamp(payload []byte) (time.Time, error) {
	var envelope struct {
		SignatureTimestamp string `json:"signature_timestamp"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", ErrMissingTimestamp, err)
	}
	if envelope.SignatureTimestamp == "" {
		return time.Time{}, ErrMissingTimestamp
	}

	parsed, err := time.Parse(time.RFC3339, envelope.SignatureTimestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", ErrMissingTimestamp, err)
	}
	return parsed, nil
}
