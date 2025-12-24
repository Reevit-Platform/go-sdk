package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"strings"
)

// Metadata keys expected by Reevit webhooks.
const (
	MetadataOrgID        = "org_id"
	MetadataConnectionID = "connection_id"
	MetadataPaymentID    = "payment_id"
)

// BuildMetadata returns the canonical metadata map expected by Reevit webhooks.
func BuildMetadata(orgID, connectionID, paymentID string) map[string]string {
	meta := make(map[string]string, 3)
	if trimmed := strings.TrimSpace(orgID); trimmed != "" {
		meta[MetadataOrgID] = trimmed
	}
	if trimmed := strings.TrimSpace(connectionID); trimmed != "" {
		meta[MetadataConnectionID] = trimmed
	}
	if trimmed := strings.TrimSpace(paymentID); trimmed != "" {
		meta[MetadataPaymentID] = trimmed
	}
	return meta
}

// SignPaystack returns the X-Paystack-Signature header (HMAC SHA512 of the raw body).
func SignPaystack(body []byte, secretKey string) string {
	return signHex(body, secretKey, sha512.New)
}

// SignHubtel returns the X-Hubtel-Signature header (HMAC SHA256 of the raw body).
func SignHubtel(body []byte, clientSecret string) string {
	return signHex(body, clientSecret, sha256.New)
}

// SignPolar returns the X-Polar-Signature header (HMAC SHA256 of the raw body).
func SignPolar(body []byte, clientSecret string) string {
	return signHex(body, clientSecret, sha256.New)
}

// FlutterwaveHash simply returns the verif-hash header value (no hashing required).
func FlutterwaveHash(secretHash string) string {
	return strings.TrimSpace(secretHash)
}

func signHex(body []byte, secret string, factory func() hash.Hash) string {
	secret = strings.TrimSpace(secret)
	if len(body) == 0 || secret == "" {
		return ""
	}
	mac := hmac.New(factory, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
