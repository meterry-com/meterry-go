package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/meterry-com/meterry-go/pkg/types"
	"github.com/shopspring/decimal"
)

func TestManagerUsageControlLimitCommands(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, method, path, body, response string
		call                               func(*Client) (any, error)
		want                               any
	}{
		{
			name: "unlimited cap", method: http.MethodPatch, path: "/usage-control-rules/limits/racl_a",
			body: `{"limit_amount":"-1"}`, response: `{"id":"racl_a","limit_amount":"-1","period_type":"calendar","init_amount":"12.345678"}`,
			call: func(c *Client) (any, error) {
				cap := decimal.NewFromInt(-1)
				return c.Manager.UpdateUsageControlLimit(ctx, "racl_a", types.UpdateUsageControlLimitRequest{LimitAmount: &cap})
			},
			want: &types.RuleAccountLimit{ID: "racl_a", LimitAmount: decimal.NewFromInt(-1), PeriodType: "calendar", InitAmount: decimal.RequireFromString("12.345678")},
		},
		{
			name: "zero cap and explicit empty fields", method: http.MethodPatch, path: "/usage-control-rules/limits/racl_a",
			body:     `{"limit_amount":"0","period":"month","period_type":"calendar","timezone":"Asia/Singapore","action":"","webhook_event":""}`,
			response: `{"id":"racl_a"}`,
			call: func(c *Client) (any, error) {
				cap := decimal.Zero
				period, periodType, timezone, empty := "month", "calendar", "Asia/Singapore", ""
				return c.Manager.UpdateUsageControlLimit(ctx, "racl_a", types.UpdateUsageControlLimitRequest{LimitAmount: &cap, Period: &period, PeriodType: &periodType, Timezone: &timezone, Action: &empty, WebhookEvent: &empty})
			},
			want: &types.RuleAccountLimit{ID: "racl_a"},
		},
		{
			name: "reset overrides scope and target", method: http.MethodPost, path: "/usage-limit-resets",
			body:     `{"scope":"usage_control_limit","target_id":"racl_a","reason":"support","idempotency_key":"reset"}`,
			response: `{"id":"reset_a","scope":"usage_control_limit","target_id":"racl_a","generation":2,"effective_occurred_at":100,"reason":"support"}`,
			call: func(c *Client) (any, error) {
				return c.Manager.ResetUsageControlLimit(ctx, "racl_a", types.CreateUsageLimitResetRequest{Scope: types.UsageLimitResetScopeProject, TargetID: "other", Reason: "support", IdempotencyKey: "reset"})
			},
			want: &types.UsageLimitResetOperation{ID: "reset_a", Scope: types.UsageLimitResetScopeLimit, TargetID: "racl_a", Generation: 2, EffectiveOccurredAt: 100, Reason: "support"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, status := range []int{http.StatusOK, http.StatusConflict} {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != tc.method || r.URL.Path != "/gateway/v1/projects/p"+tc.path || r.URL.RawQuery != "" {
						t.Errorf("request = %s %s", r.Method, r.URL.RequestURI())
					}
					if r.Header.Get("Authorization") != "Bearer key" || r.Header.Get("X-Tenant-ID") != "" {
						t.Errorf("headers = %v", r.Header)
					}
					if tc.body != "" {
						var got, want map[string]any
						if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
							t.Error(err)
						}
						if err := json.Unmarshal([]byte(tc.body), &want); err != nil {
							t.Error(err)
						}
						if !reflect.DeepEqual(got, want) {
							t.Errorf("body = %#v; want %#v", got, want)
						}
					} else if r.ContentLength != 0 {
						t.Errorf("GET body length = %d", r.ContentLength)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					if status == http.StatusConflict {
						_, _ = w.Write([]byte(`{"error":"conflict"}`))
						return
					}
					_, _ = w.Write([]byte(tc.response))
				}))
				client := MustNewClient(Config{BaseURL: server.URL + "/gateway/", ProjectID: "p", APIKey: "key"})
				got, err := tc.call(client)
				server.Close()
				if status == http.StatusConflict {
					var apiErr *APIError
					if !errors.As(err, &apiErr) || apiErr.StatusCode != status || apiErr.Message != "conflict" {
						t.Fatalf("error = %v", err)
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				// Compare wire values so decimal scale and zero representations are equivalent.
				gotJSON, err := json.Marshal(got)
				if err != nil {
					t.Fatal(err)
				}
				wantJSON, err := json.Marshal(tc.want)
				if err != nil {
					t.Fatal(err)
				}
				if string(gotJSON) != string(wantJSON) {
					t.Fatalf("response = %s; want %s", gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestManagerUsageControlCarriesPeriodTypeAndInitialAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Limits []map[string]any `json:"limits"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if len(body.Limits) != 1 || body.Limits[0]["period_type"] != "calendar" || body.Limits[0]["init_amount"] != "12.345678" {
			t.Errorf("limit configuration lost in transport: %#v", body.Limits)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"limits":[{"period_type":"calendar","init_amount":"12.345678"}]}`))
	}))
	defer server.Close()
	client := MustNewClient(Config{BaseURL: server.URL, ProjectID: "p", APIKey: "key"})
	response, err := client.Manager.CreateUsageControlRule(context.Background(), types.CreateUsageControlRuleRequest{
		BillingAccountID: "a", SubjectType: "user", SubjectID: "u",
		Limits: []types.CreateRuleAccountLimitRequest{{Period: "month", PeriodType: "calendar", InitAmount: decimal.RequireFromString("12.345678"), LimitAmount: decimal.NewFromInt(100)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Limits) != 1 || response.Limits[0].PeriodType != "calendar" || !response.Limits[0].InitAmount.Equal(decimal.RequireFromString("12.345678")) {
		t.Fatalf("limit response lost fields: %#v", response)
	}
}
