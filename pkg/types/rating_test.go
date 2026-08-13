package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestUsageChargeDetailOmitsRuntimeProjectionFieldsFromJSON(t *testing.T) {
	detail := UsageChargeDetail{
		Kind:            UsageChargeDetailKindQuantity,
		LimitDimensions: map[string]string{"model": "gpt-5"},
		LimitItems:      []ChargeItem{{Metric: "input_tokens", Quantity: 100}},
		Amount:          decimal.NewFromInt(1),
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"kind", "limit_dimensions", "limit_items"} {
		if strings.Contains(string(payload), "\""+field+"\"") {
			t.Fatalf("runtime field %q leaked into JSON: %s", field, payload)
		}
	}
}
