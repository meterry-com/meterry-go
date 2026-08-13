package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meterry-com/meterry-go/pkg/types"
)

const (
	WebhookIDHeader         = "X-Billing-Webhook-ID"
	WebhookDeliveryIDHeader = "X-Billing-Webhook-Delivery-ID"
	WebhookTimestampHeader  = "X-Billing-Webhook-Timestamp"
	WebhookSignatureHeader  = "X-Billing-Webhook-Signature"

	DefaultWebhookSignatureTolerance = 5 * time.Minute
)

var (
	ErrWebhookMissingSecret     = errors.New("webhook signing secret is required")
	ErrWebhookNilRequest        = errors.New("webhook request is nil")
	ErrWebhookMissingTimestamp  = errors.New("webhook timestamp header is required")
	ErrWebhookInvalidTimestamp  = errors.New("webhook timestamp header is invalid")
	ErrWebhookMissingSignature  = errors.New("webhook signature header is required")
	ErrWebhookInvalidSignature  = errors.New("webhook signature header is invalid")
	ErrWebhookSignatureMismatch = errors.New("webhook signature mismatch")
	ErrWebhookTimestampExpired  = errors.New("webhook timestamp is outside tolerance")
)

func VerifyWebhookRequest(req *http.Request, body []byte, secret string) error {
	return VerifyWebhookRequestWithTolerance(req, body, secret, DefaultWebhookSignatureTolerance)
}

func VerifyWebhookRequestWithTolerance(req *http.Request, body []byte, secret string, tolerance time.Duration) error {
	if req == nil {
		return ErrWebhookNilRequest
	}
	timestamp, err := webhookTimestamp(req.Header.Get(WebhookTimestampHeader))
	if err != nil {
		return err
	}
	if err = VerifyWebhookSignature(secret, req.Header.Get(WebhookSignatureHeader), timestamp, body); err != nil {
		return err
	}
	return verifyWebhookTimestamp(timestamp, time.Now(), tolerance)
}

func VerifyWebhookSignature(secret, signature string, timestamp int64, body []byte) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ErrWebhookMissingSecret
	}
	received, err := webhookSignatureBytes(signature)
	if err != nil {
		return err
	}
	expected := webhookExpectedSignature(secret, timestamp, body)
	if !hmac.Equal(received, expected) {
		return ErrWebhookSignatureMismatch
	}
	return nil
}

func webhookTimestamp(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ErrWebhookMissingTimestamp
	}
	timestamp, err := strconv.ParseInt(value, 10, 64)
	if err != nil || timestamp <= 0 {
		return 0, ErrWebhookInvalidTimestamp
	}
	return timestamp, nil
}

func webhookSignatureBytes(signature string) ([]byte, error) {
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return nil, ErrWebhookMissingSignature
	}
	for _, part := range strings.Split(signature, ",") {
		part = strings.TrimSpace(part)
		value, ok := strings.CutPrefix(part, "v1=")
		if !ok {
			continue
		}
		raw, err := hex.DecodeString(value)
		if err != nil || len(raw) != sha256.Size {
			return nil, ErrWebhookInvalidSignature
		}
		return raw, nil
	}
	return nil, ErrWebhookInvalidSignature
}

func webhookExpectedSignature(secret string, timestamp int64, body []byte) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(types.WebhookSignaturePayload(timestamp, body)))
	return mac.Sum(nil)
}

func verifyWebhookTimestamp(timestamp int64, now time.Time, tolerance time.Duration) error {
	if tolerance <= 0 {
		return nil
	}
	deliveryTime := time.Unix(timestamp, 0)
	if deliveryTime.Before(now.Add(-tolerance)) || deliveryTime.After(now.Add(tolerance)) {
		return ErrWebhookTimestampExpired
	}
	return nil
}
