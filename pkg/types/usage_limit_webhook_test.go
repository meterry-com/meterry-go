package types

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestUsageLimitWebhookPayload(t *testing.T) {
	for _, eventType := range []WebhookEventType{WebhookEventTypeUsageLimitExhausted, WebhookEventTypeUsageLimitRecovered} {
		t.Run(string(eventType), func(t *testing.T) {
			body := `{"event_type":"` + string(eventType) + `","subject":{"type":"org_user","id":"member_001"},"limit":{"id":"racl_key_a","rule_kind":"subject_route","dimension_key":"api_key_id","dimension_value":"key_a","used_amount":"123.456789","previous_used_amount":"120.1","limit_amount":"100","remaining_amount":"-23.456789","next_reset_at":200},"recovery":{"reason":"window_reset","recovered_at":100}}`
			var payload UsageLimitWebhookPayload
			if err := json.Unmarshal([]byte(body), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.EventType != eventType || payload.Subject.ID != "member_001" || payload.Limit.DimensionKey != "api_key_id" || payload.Limit.DimensionValue != "key_a" || payload.Limit.RuleKind != UsageLimitRuleKindSubjectRoute {
				t.Fatalf("limit identity lost: %#v", payload)
			}
			if !payload.Limit.UsedAmount.Equal(decimal.RequireFromString("123.456789")) || !payload.Limit.PreviousUsedAmount.Equal(decimal.RequireFromString("120.1")) || !payload.Limit.RemainingAmount.Equal(decimal.RequireFromString("-23.456789")) || payload.Limit.NextResetAt != 200 {
				t.Fatalf("limit amounts lost: %#v", payload.Limit)
			}
			if payload.Recovery == nil || payload.Recovery.Reason != "window_reset" || payload.Recovery.RecoveredAt != 100 {
				t.Fatalf("recovery = %#v", payload.Recovery)
			}
		})
	}
	var exhausted UsageLimitWebhookPayload
	if err := json.Unmarshal([]byte(`{"event_type":"usage_limit.exhausted","limit":{"id":"racl_a"}}`), &exhausted); err != nil {
		t.Fatal(err)
	}
	if exhausted.Recovery != nil {
		t.Fatalf("unexpected recovery = %#v", exhausted.Recovery)
	}
}
