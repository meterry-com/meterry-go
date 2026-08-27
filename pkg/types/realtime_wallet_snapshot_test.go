package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestRealtimeWalletPublicSnapshotsRedactImplementationFields(t *testing.T) {
	amountJSON, err := json.Marshal(VirtualWalletAmountSnapshot{
		TenantID: "tenant_001", AccountID: "acct_001", Currency: "USD", Balance: decimal.NewFromInt(10),
	})
	if err != nil {
		t.Fatal(err)
	}
	limitsJSON, err := json.Marshal(VirtualWalletLimitsSnapshot{
		TenantID: "tenant_001", AccountID: "acct_001", SubjectType: "user", SubjectID: "user_001", Currency: "USD",
		Limits: []VirtualWalletLimitSnapshot{{MetricScope: "tokens", Period: "month", UsageVersion: 42}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{string(amountJSON), string(limitsJSON)} {
		if strings.Contains(raw, `"usage_version"`) || strings.Contains(raw, `"scale"`) {
			t.Fatalf("public realtime wallet payload leaked implementation fields: %s", raw)
		}
	}
}

func TestSubjectRealtimeLimitsPublicSnapshotRedactsImplementationFields(t *testing.T) {
	raw, err := json.Marshal(SubjectRealtimeLimitsSnapshot{
		TenantID:    "tenant_001",
		ProjectID:   "proj_001",
		AccountID:   "acct_001",
		SubjectType: "user",
		SubjectID:   "user_001",
		Limits: []SubjectRealtimeLimitItem{{
			RuleAccountID:   "asub_001",
			SubjectType:     "user",
			SubjectID:       "user_001",
			DimensionKey:    "api_key_id",
			DimensionValue:  "key_001",
			Currency:        "USD",
			MetricScope:     "tokens",
			SelectorType:    "metric",
			SelectorValue:   "input_tokens",
			Measure:         "quantity",
			Period:          "month",
			WindowID:        "2026-08",
			NextResetAt:     1234567890,
			UsedAmount:      decimal.RequireFromString("1"),
			LimitAmount:     decimal.RequireFromString("10"),
			RemainingAmount: decimal.RequireFromString("9"),
			Exceeded:        false,
			Action:          "block",
			WebhookEvent:    "usage.monthly_tokens_exceeded",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"usage_version"`) || strings.Contains(string(raw), `"scale"`) {
		t.Fatalf("public subject realtime snapshot leaked implementation fields: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"next_reset_at":1234567890`) {
		t.Fatalf("subject realtime snapshot omitted next_reset_at: %s", string(raw))
	}
}
