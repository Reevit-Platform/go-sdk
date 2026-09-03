package webhooks

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Canonical known-answer vector shared with the Node, Python and PHP SDK
// tests. Any change to the signing scheme must move all of them together.
const (
	vectorSecret    = "whsec_test_2x9aBcDeFgHiJkLmNoPqRsTuVwXyZ012"
	vectorBody      = `{"event":"payment.updated","org_id":"org_123","signature_timestamp":"2026-06-13T12:00:00Z","data":{"id":"pay_abc","status":"succeeded"}}`
	vectorSignature = "sha256=8fed6e24bd1c97ac5634ec88081a8299107706bf488290513bc3e5c5340e1950"
)

func TestSignMatchesCrossLanguageVector(t *testing.T) {
	require.Equal(t, vectorSignature, Sign([]byte(vectorBody), vectorSecret))
}

func TestVerifyAcceptsTheVector(t *testing.T) {
	require.True(t, Verify([]byte(vectorBody), vectorSignature, vectorSecret))
}

func TestVerifyRejectsTamperedInput(t *testing.T) {
	tampered := []byte(`{"event":"payment.updated","org_id":"org_123","signature_timestamp":"2026-06-13T12:00:00Z","data":{"id":"pay_abc","status":"failed"}}`)

	require.False(t, Verify(tampered, vectorSignature, vectorSecret), "tampered body")
	require.False(t, Verify([]byte(vectorBody), vectorSignature, "whsec_wrong"), "wrong secret")
	require.False(t, Verify([]byte(vectorBody), "", vectorSecret), "empty signature")
	require.False(t, Verify([]byte(vectorBody), vectorSignature, ""), "empty secret")
	require.False(t, Verify([]byte(vectorBody), "sha256=deadbeef", vectorSecret), "short signature")
	require.False(t, Verify([]byte(vectorBody), vectorSignature[len(SignaturePrefix):], vectorSecret), "unprefixed signature")
}

func TestVerifyWithToleranceUsesTheSignedTimestamp(t *testing.T) {
	fresh := fmt.Sprintf(`{"event":"payment.updated","signature_timestamp":%q}`, time.Now().UTC().Format(time.RFC3339))
	signature := Sign([]byte(fresh), vectorSecret)

	require.NoError(t, VerifyWithTolerance([]byte(fresh), signature, vectorSecret, time.Time{}, 5*time.Minute))

	// The vector's timestamp is a fixed date, so it is always outside a
	// five-minute window.
	err := VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Time{}, 5*time.Minute)
	require.ErrorIs(t, err, ErrTimestampOutsideTolerance)
}

func TestVerifyWithToleranceAcceptsAnExplicitSignedAt(t *testing.T) {
	require.NoError(t, VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Now().Add(-time.Minute), 5*time.Minute))

	err := VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Now().Add(-time.Hour), 5*time.Minute)
	require.ErrorIs(t, err, ErrTimestampOutsideTolerance)

	// A timestamp far in the future is just as suspect as a stale one.
	err = VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Now().Add(time.Hour), 5*time.Minute)
	require.ErrorIs(t, err, ErrTimestampOutsideTolerance)
}

func TestVerifyWithToleranceChecksTheSignatureFirst(t *testing.T) {
	err := VerifyWithTolerance([]byte(vectorBody), vectorSignature, "whsec_wrong", time.Now(), time.Minute)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestVerifyWithToleranceRequiresATimestamp(t *testing.T) {
	body := []byte(`{"event":"payment.updated"}`)
	err := VerifyWithTolerance(body, Sign(body, vectorSecret), vectorSecret, time.Time{}, time.Minute)
	require.ErrorIs(t, err, ErrMissingTimestamp)

	unparsable := []byte(`{"signature_timestamp":"13/06/2026"}`)
	err = VerifyWithTolerance(unparsable, Sign(unparsable, vectorSecret), vectorSecret, time.Time{}, time.Minute)
	require.ErrorIs(t, err, ErrMissingTimestamp)
}

func TestVerifyWithToleranceDefaultsToFiveMinutes(t *testing.T) {
	require.NoError(t, VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Now().Add(-4*time.Minute), 0))

	err := VerifyWithTolerance([]byte(vectorBody), vectorSignature, vectorSecret, time.Now().Add(-6*time.Minute), 0)
	require.True(t, errors.Is(err, ErrTimestampOutsideTolerance))
}
