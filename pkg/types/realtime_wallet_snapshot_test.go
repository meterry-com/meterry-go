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
