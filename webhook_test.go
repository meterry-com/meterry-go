package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestVerifyWebhookSignature(t *testing.T) {
	body := []byte(`{"id":"wh_evt_001","type":"usage_limit.exhausted"}`)
	timestamp := int64(1783562400)
	signature := testWebhookSignature("whsec_test", timestamp, body)

	if err := VerifyWebhookSignature("whsec_test", signature, timestamp, body); err != nil {
		t.Fatal(err)
	}
	if err := VerifyWebhookSignature("wrong", signature, timestamp, body); !errors.Is(err, ErrWebhookSignatureMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookSignatureMismatch)
	}
	if err := VerifyWebhookSignature("whsec_test", "v2=abc", timestamp, body); !errors.Is(err, ErrWebhookInvalidSignature) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookInvalidSignature)
	}
}

func TestVerifyWebhookRequest(t *testing.T) {
	body := []byte(`{"id":"wh_evt_001","type":"wallet.insufficient_balance"}`)
	timestamp := time.Now().Unix()
	req, err := http.NewRequest(http.MethodPost, "https://example.test/webhooks", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(WebhookTimestampHeader, " "+testFormatInt(timestamp)+" ")
	req.Header.Set(WebhookSignatureHeader, "v0=ignored, "+testWebhookSignature("whsec_test", timestamp, body))

	if err = VerifyWebhookRequest(req, body, "whsec_test"); err != nil {
		t.Fatal(err)
	}
	oldTimestamp := timestamp - 3600
	req.Header.Set(WebhookTimestampHeader, testFormatInt(oldTimestamp))
	req.Header.Set(WebhookSignatureHeader, testWebhookSignature("whsec_test", oldTimestamp, body))
	if err = VerifyWebhookRequest(req, body, "whsec_test"); !errors.Is(err, ErrWebhookTimestampExpired) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookTimestampExpired)
	}
	if err = VerifyWebhookRequestWithTolerance(req, body, "whsec_test", 0); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWebhookRequestRequiresHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.test/webhooks", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = VerifyWebhookRequest(req, []byte(`{}`), "whsec_test"); !errors.Is(err, ErrWebhookMissingTimestamp) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookMissingTimestamp)
	}
	req.Header.Set(WebhookTimestampHeader, "not-a-number")
	if err = VerifyWebhookRequest(req, []byte(`{}`), "whsec_test"); !errors.Is(err, ErrWebhookInvalidTimestamp) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookInvalidTimestamp)
	}
	req.Header.Set(WebhookTimestampHeader, testFormatInt(time.Now().Unix()))
	if err = VerifyWebhookRequest(req, []byte(`{}`), "whsec_test"); !errors.Is(err, ErrWebhookMissingSignature) {
		t.Fatalf("error = %v, want %v", err, ErrWebhookMissingSignature)
	}
}

func testWebhookSignature(secret string, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(types.WebhookSignaturePayload(timestamp, body)))
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func testFormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
