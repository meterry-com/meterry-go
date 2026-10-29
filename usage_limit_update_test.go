package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/meterry-com/meterry-go/pkg/types"
	"github.com/shopspring/decimal"
)

func TestUpdateRuleAccountLimitContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.RequestURI() != "/gateway/v1/projects/proj_001/usage-control-rules/limits/racl_001" {
			t.Errorf("request = %s %s", r.Method, r.URL.RequestURI())
		}
		if r.Header.Get("Authorization") != "Bearer tenant_key" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("X-Tenant-ID") != "" {
			t.Errorf("headers = %v", r.Header)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(body, map[string]any{"limit_amount": "20.123456"}) {
			t.Errorf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		// The server returns one realtime limit directly, without a wrapper.
		_, _ = w.Write([]byte(`{"id":"racl_001","rule_account_id":"asub_001","subject_type":"user","subject_id":"u1","dimension_key":"api_key","dimension_value":"k1","currency":"USD","metric_scope":"amount","selector_type":"all","measure":"amount","period":"month","period_type":"current","window_id":"1792893600","next_reset_at":1795572000,"used_amount":"30.123456","limit_amount":"20.123456","remaining_amount":"-10","exceeded":true,"action":"block","webhook_event":"usage.exceeded"}`))
	}))
	defer server.Close()
	client := MustNewClient(Config{BaseURL: server.URL + "/gateway", ProjectID: "proj_001", APIKey: "tenant_key"})
	got, err := client.Manager.UpdateRuleAccountLimit(context.Background(), "racl_001", types.UpdateRuleAccountLimitRequest{LimitAmount: decimal.RequireFromString("20.123456")})
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("missing realtime limit")
	}
	want := types.SubjectRealtimeLimitItem{
		ID: "racl_001", RuleAccountID: "asub_001", SubjectType: "user", SubjectID: "u1", DimensionKey: "api_key", DimensionValue: "k1", Currency: "USD", MetricScope: "amount", SelectorType: "all", Measure: "amount", Period: "month", PeriodType: "current", WindowID: "1792893600", NextResetAt: 1795572000,
		UsedAmount: decimal.RequireFromString("30.123456"), LimitAmount: decimal.RequireFromString("20.123456"), RemainingAmount: decimal.NewFromInt(-10), Exceeded: true, Action: "block", WebhookEvent: "usage.exceeded",
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("realtime limit = %#v; want %#v", *got, want)
	}
}

func TestUpdateRuleAccountLimitReturnsAPIError(t *testing.T) {
	for _, tc := range []struct {
		status  int
		message string
	}{
		{400, "limit_amount must be greater than zero"},
		{401, "unauthorized"},
		{403, "forbidden"},
		{404, "active custom usage limit not found"},
		{503, "failed to update realtime limit amount"},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": tc.message})
			}))
			defer server.Close()
			client := MustNewClient(Config{BaseURL: server.URL, ProjectID: "proj", APIKey: "key"})
			_, err := client.Manager.UpdateRuleAccountLimit(context.Background(), "racl", types.UpdateRuleAccountLimitRequest{LimitAmount: decimal.NewFromInt(20)})
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status || apiErr.Message != tc.message {
				t.Fatalf("error = %v; want APIError(%d, %q)", err, tc.status, tc.message)
			}
		})
	}
}

func TestUpdateRuleAccountLimitPublicJSON(t *testing.T) {
	raw, err := json.Marshal(types.UpdateRuleAccountLimitRequest{LimitAmount: decimal.NewFromInt(20)})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"limit_amount":"20"}` {
		t.Fatalf("request = %s", raw)
	}
	var item types.SubjectRealtimeLimitItem
	if err := json.Unmarshal([]byte(`{"id":"r1","period_type":"calendar","used_amount":"30","limit_amount":"20","remaining_amount":"-10","exceeded":true,"usage_version":1,"scale":6,"storage_key":"internal","limit_units":20000000}`), &item); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"period_type":"calendar"`) {
		t.Fatalf("missing period type: %s", raw)
	}
	for _, hidden := range []string{"usage_version", "scale", "storage_key", "limit_units"} {
		if strings.Contains(string(raw), `"`+hidden+`"`) {
			t.Fatalf("public snapshot leaked %s: %s", hidden, raw)
		}
	}
}
