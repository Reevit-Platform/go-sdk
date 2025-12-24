# Reevit Go SDK

The official Go SDK for [Reevit](https://reevit.io) — a unified payment orchestration platform for Africa.

[![Go Reference](https://pkg.go.dev/badge/github.com/Reevit-Platform/go-sdk.svg)](https://pkg.go.dev/github.com/Reevit-Platform/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Installation

```bash
go get github.com/Reevit-Platform/go-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	reevit "github.com/Reevit-Platform/go-sdk"
)

func main() {
	// Initialize the client
	client := reevit.NewClient("pfk_live_xxx")

	// Create a payment intent
	payment, err := client.Payments.CreateIntent(context.Background(), &reevit.PaymentIntentRequest{
		Amount:   45000,
		Currency: "GHS",
		Method:   "momo",
		Country:  "GH",
		Metadata: map[string]interface{}{
			"order_id": "12345",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Payment created: %s (Status: %s)\n", payment.ID, payment.Status)
}
```

## Services

- **Payments**: `client.Payments` (CreateIntent, Get, List, Refund)
- **Connections**: `client.Connections` (Create, List, Test)
- **Subscriptions**: `client.Subscriptions` (Create, List)
- **Fraud**: `client.Fraud` (Get, Update)

---

## Webhook Verification

Reevit sends webhooks to notify your application of payment events. Always verify webhook signatures.

### Understanding Webhooks

There are **two types of webhooks** in Reevit:

1. **Inbound Webhooks (PSP → Reevit)**: Webhooks from payment providers (Paystack, Flutterwave, etc.) to Reevit. Configure these in the PSP dashboard. Reevit handles them automatically.

2. **Outbound Webhooks (Reevit → Your App)**: Webhooks from Reevit to your application. Configure in Reevit Dashboard and create a handler in your app.

### Signature Format

- **Header**: `X-Reevit-Signature: sha256=<hex-signature>`
- **Signature**: `HMAC-SHA256(request_body, signing_secret)`

### Getting Your Signing Secret

1. Go to **Reevit Dashboard > Developers > Webhooks**
2. Configure your webhook endpoint URL
3. Copy the signing secret (starts with `whsec_`)
4. Set environment variable: `REEVIT_WEBHOOK_SECRET=whsec_xxx...`

### Webhook Handler Example

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// PaymentData represents payment event data
type PaymentData struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	Amount     int64             `json:"amount"`
	Currency   string            `json:"currency"`
	Provider   string            `json:"provider"`
	CustomerID string            `json:"customer_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// SubscriptionData represents subscription event data
type SubscriptionData struct {
	ID            string `json:"id"`
	CustomerID    string `json:"customer_id"`
	PlanID        string `json:"plan_id"`
	Status        string `json:"status"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Interval      string `json:"interval"`
	NextRenewalAt string `json:"next_renewal_at,omitempty"`
}

// WebhookPayload represents the webhook event structure
type WebhookPayload struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	OrgID     string          `json:"org_id"`
	CreatedAt string          `json:"created_at"`
	Data      json.RawMessage `json:"data,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// VerifySignature verifies the webhook signature using HMAC-SHA256
func VerifySignature(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	received := signature[7:] // Remove "sha256=" prefix

	return hmac.Equal([]byte(received), []byte(expected))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read the raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get signature and secret
	signature := r.Header.Get("X-Reevit-Signature")
	secret := os.Getenv("REEVIT_WEBHOOK_SECRET")

	// Verify signature (required in production)
	if secret != "" && !VerifySignature(body, signature, secret) {
		log.Println("[Webhook] Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse the event
	var event WebhookPayload
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[Webhook] Received: %s (%s)", event.Type, event.ID)

	// Handle different event types
	switch event.Type {
	// Test event
	case "reevit.webhook.test":
		log.Printf("[Webhook] Test received: %s", event.Message)

	// Payment events
	case "payment.succeeded":
		var data PaymentData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handlePaymentSucceeded(data)
		}

	case "payment.failed":
		var data PaymentData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handlePaymentFailed(data)
		}

	case "payment.refunded":
		var data PaymentData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handlePaymentRefunded(data)
		}

	case "payment.pending":
		var data PaymentData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			log.Printf("[Webhook] Payment pending: %s", data.ID)
		}

	// Subscription events
	case "subscription.created":
		var data SubscriptionData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handleSubscriptionCreated(data)
		}

	case "subscription.renewed":
		var data SubscriptionData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handleSubscriptionRenewed(data)
		}

	case "subscription.canceled":
		var data SubscriptionData
		if err := json.Unmarshal(event.Data, &data); err == nil {
			handleSubscriptionCanceled(data)
		}

	default:
		log.Printf("[Webhook] Unhandled event: %s", event.Type)
	}

	// Acknowledge receipt
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"received": true})
}

// Payment handlers
func handlePaymentSucceeded(data PaymentData) {
	orderID := data.Metadata["order_id"]
	log.Printf("[Webhook] Payment succeeded: %s for order %s", data.ID, orderID)

	// TODO: Implement your business logic
	// - Update order status to "paid"
	// - Send confirmation email to customer
	// - Trigger fulfillment process
}

func handlePaymentFailed(data PaymentData) {
	log.Printf("[Webhook] Payment failed: %s", data.ID)

	// TODO: Implement your business logic
	// - Update order status to "payment_failed"
	// - Send notification to customer
	// - Allow retry
}

func handlePaymentRefunded(data PaymentData) {
	orderID := data.Metadata["order_id"]
	log.Printf("[Webhook] Payment refunded: %s for order %s", data.ID, orderID)

	// TODO: Implement your business logic
	// - Update order status to "refunded"
	// - Restore inventory if applicable
}

// Subscription handlers
func handleSubscriptionCreated(data SubscriptionData) {
	log.Printf("[Webhook] Subscription created: %s for customer %s", data.ID, data.CustomerID)

	// TODO: Implement your business logic
	// - Grant access to subscription features
	// - Send welcome email
}

func handleSubscriptionRenewed(data SubscriptionData) {
	log.Printf("[Webhook] Subscription renewed: %s", data.ID)

	// TODO: Implement your business logic
	// - Extend access period
	// - Send renewal confirmation
}

func handleSubscriptionCanceled(data SubscriptionData) {
	log.Printf("[Webhook] Subscription canceled: %s", data.ID)

	// TODO: Implement your business logic
	// - Revoke access at end of billing period
	// - Send cancellation confirmation
}

func main() {
	http.HandleFunc("/webhooks/reevit", webhookHandler)
	log.Println("Webhook server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Using the webhooks Subpackage

For convenience, use the `webhooks` subpackage:

```go
import "github.com/Reevit-Platform/go-sdk/webhooks"

// Verify signature
isValid := webhooks.VerifySignature(body, signature, secret)
```

---

## Environment Variables

```bash
export REEVIT_API_KEY=pfk_live_xxx
export REEVIT_ORG_ID=org_xxx
export REEVIT_WEBHOOK_SECRET=whsec_xxx  # Get from Dashboard > Developers > Webhooks
```

---

## Release Notes

### v0.3.0

- Updated API client to connect to production and sandbox URLs based on API key
- Added Bearer authentication headers for secure API communication
- Added Get, Confirm, and Cancel methods to the payments service
- Removed orgID parameter from client initialization (simplified API)
- Added .gitignore file to exclude unnecessary files from version control
- Updated README.md with corrected quick start example

---

## Support

- **Documentation**: [https://docs.reevit.io](https://docs.reevit.io)
- **GitHub Issues**: [https://github.com/Reevit-Platform/backend/issues](https://github.com/Reevit-Platform/backend/issues)
- **Email**: support@reevit.io

## License

MIT License - see [LICENSE](LICENSE) for details.
