package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestManagerUsesProjectPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Header.Get("Authorization") != "Bearer service_key" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"acct_001","tenant_id":"t_001","status":"active"}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{
		ProjectID: "proj_001",
		BaseURL:   server.URL,
		APIKey:    "service_key",
	})
	account, err := client.Manager.CreateAccount(context.Background(), types.CreateAccountRequest{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/projects/proj_001/accounts" {
		t.Fatalf("path = %q", gotPath)
	}
	if account.ID != "acct_001" {
		t.Fatalf("account = %#v", account)
	}
}

func TestNewClientRequiresScopeAndCredentials(t *testing.T) {
	valid := Config{ProjectID: "proj_001", BaseURL: "https://api.meterry.com", APIKey: "service_key"}
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"project", func(c *Config) { c.ProjectID = " \t" }, "project id is required"},
		{"api key", func(c *Config) { c.APIKey = " \t" }, "api key is required"},
		{"base url", func(c *Config) { c.BaseURL = " \t" }, "base url is required"},
		{"invalid base url", func(c *Config) { c.BaseURL = "://" }, "parse billing sdk base url"},
		{"relative base url", func(c *Config) { c.BaseURL = "/api" }, "base url must include scheme and host"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := valid
			tt.edit(&config)
			client, err := NewClient(config)
			if client != nil || err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewClient = %v, %v; want %q", client, err, tt.want)
			}
		})
	}
}

func TestClientsUseConfiguredProjectAndCredentials(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path+" "+r.Header.Get("Authorization"))
		if r.Header.Get("X-Tenant-ID") != "" {
			t.Error("tenant header is not supported by the API")
		}
		if r.Header.Get("User-Agent") != "test-sdk" {
			t.Errorf("user agent = %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[{"dimensions":{"bucket":"1785000000"},"measures":{"quantity":"2"}}]}`))
	}))
	defer server.Close()

	for _, project := range []string{"proj_alpha", "proj_beta"} {
		config := Config{
			ProjectID: " " + project + " ",
			BaseURL:   " " + server.URL + "/gateway/ ", APIKey: " key_" + project + " ",
			HTTPClient: server.Client(), UserAgent: "test-sdk",
		}
		client := MustNewClient(config)
		if client.httpClient != config.HTTPClient {
			t.Fatalf("client config not retained: %#v", client)
		}
		config.ProjectID = "proj_changed"
		ctx := context.Background()
		if _, err := client.Manager.CreateExtractorRuleSet(ctx, types.CreateExtractorRuleSetRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Query.Report(ctx, "report_001", types.UsageAnalyticsQueryRequest{Metrics: []string{"tokens"}, Measures: []string{"quantity"}}); err != nil {
			t.Fatal(err)
		}
		if series, err := client.Query.Chart(ctx, ChartQuery{Metrics: []string{"tokens"}}); err != nil || len(series) != 1 {
			t.Fatalf("Chart = %#v, %v", series, err)
		}
		if _, err := client.Query.ListProjectUsageEvents(ctx, types.ListUsageEventLogsRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Query.ListProjectUsageEventDetails(ctx, types.ListUsageEventLogsRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Query.ListUsageLogs(ctx, types.ListUsageLogsRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Ingest.IngestRawEvent(ctx, "ers_001", types.IngestEventRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Ingest.BatchIngestRawEvents(ctx, types.BatchIngestEventsRequest{}); err != nil {
			t.Fatal(err)
		}
	}
	want := make([]string, 0)
	for _, project := range []string{"proj_alpha", "proj_beta"} {
		for _, route := range []struct{ method, suffix string }{
			{http.MethodPost, "extractor-rule-sets"},
			{http.MethodPost, "reports/report_001/query"},
			{http.MethodPost, "usage/query"},
			{http.MethodGet, "usage/events"},
			{http.MethodGet, "usage/events/details"},
			{http.MethodGet, "usage-logs"},
			{http.MethodPost, "extractor-rule-sets/ers_001/events/ingest"},
			{http.MethodPost, "events/batch-ingest"},
		} {
			want = append(want, route.method+" /gateway/v1/projects/"+project+"/"+route.suffix+" Bearer key_"+project)
		}
	}
	if strings.Join(requests, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests = %v; want %v", requests, want)
	}
}

func TestIngestRequiresExtractorRuleSetID(t *testing.T) {
	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: "http://127.0.0.1:8080", APIKey: "usage_key"})
	_, err := client.Ingest.Ingest(context.Background(), "", IngestEvent{})
	if err == nil || !strings.Contains(err.Error(), "extractor rule set id is required") {
		t.Fatalf("ingest err = %v", err)
	}
}

func TestManagerProjectScopedAccountWalletAndLimitOperations(t *testing.T) {
	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/accounts"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"acct_001","tenant_id":"t_001","project_id":"proj_001","status":"active"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/accounts"):
			_, _ = w.Write([]byte(`{"accounts":[{"id":"acct_001","tenant_id":"t_001","project_id":"proj_001","status":"active"}]}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/account-subjects"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"asub_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","subject_type":"org","subject_id":"org_001"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/account-subjects"):
			_, _ = w.Write([]byte(`{"account_subjects":[{"id":"asub_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","subject_type":"org","subject_id":"org_001"}]}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"wallet_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"0","available_balance":"0","credit_limit":"0","status":"active"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets"):
			_, _ = w.Write([]byte(`{"wallets":[{"id":"wallet_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"0","available_balance":"0","credit_limit":"0","status":"active"}]}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets/realtime/amount"):
			_, _ = w.Write([]byte(`{"tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"100","available_balance":"100","credit_limit":"0"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/subject/realtime/limits"):
			_, _ = w.Write([]byte(`{"tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","subject_type":"org","subject_id":"org_001","rule_account_id":"asub_001","limits":[{"id":"racl_001","rule_account_id":"asub_001","subject_type":"org","subject_id":"org_001","currency":"USD","metric_scope":"tokens","selector_type":"metric","selector_value":"input_tokens","measure":"quantity","period":"month","window_id":"2026-08","next_reset_at":1785686400,"used_amount":"1","limit_amount":"10","remaining_amount":"9","exceeded":false,"action":"block","webhook_event":"usage.monthly_tokens_exceeded"}]}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallet-ledger-entries"):
			_, _ = w.Write([]byte(`{"ledger_entries":[],"total_count":0,"limit":20,"offset":0}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets/credit"):
			_, _ = w.Write([]byte(`{"wallet":{"id":"wallet_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"100","available_balance":"100","credit_limit":"0","status":"active"},"ledger_entry":{"id":"ledger_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","wallet_id":"wallet_001","currency":"USD","direction":"credit","amount":"100","balance_after":"100"}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets/debit"):
			_, _ = w.Write([]byte(`{"wallet":{"id":"wallet_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"90","available_balance":"90","credit_limit":"0","status":"active"},"ledger_entry":{"id":"ledger_002","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","wallet_id":"wallet_001","currency":"USD","direction":"debit","amount":"10","balance_after":"90"}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/wallets/credit-limit"):
			_, _ = w.Write([]byte(`{"id":"wallet_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","balance":"0","available_balance":"50","credit_limit":"50","status":"active"}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/voucher-grants"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"voucher_grant":{"id":"vg_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","original_amount":"20","remaining_amount":"20","status":"active"},"ledger_entry":{"id":"ledger_003","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","direction":"credit","amount":"20","balance_after":"20"}}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/voucher-grants"):
			_, _ = w.Write([]byte(`{"voucher_grants":[{"id":"vg_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","original_amount":"20","remaining_amount":"20","status":"active"}],"total_count":1,"limit":10,"offset":0}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/voucher-grants/vg_001/revoke"):
			_, _ = w.Write([]byte(`{"voucher_grant":{"id":"vg_001","tenant_id":"t_001","project_id":"proj_001","account_id":"acct_001","currency":"USD","original_amount":"20","remaining_amount":"0","status":"revoked"},"pending_amount":"20"}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/usage-control-rules"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"rule_account":{"id":"asub_001","tenant_id":"t_001","project_id":"proj_001","billing_account_id":"acct_001","subject_type":"org","subject_id":"org_001","status":"active"},"limits":[{"id":"racl_001","tenant_id":"t_001","project_id":"proj_001","rule_kind":"subject_route","rule_id":"asub_001","currency":"USD","metric_scope":"tokens","period":"month","limit_amount":"1000","status":"active"}]}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects/proj_001/usage-control-rules"):
			_, _ = w.Write([]byte(`{"usage_control_rules":[{"rule_account":{"id":"asub_001","tenant_id":"t_001","project_id":"proj_001","billing_account_id":"acct_001","subject_type":"org","subject_id":"org_001","status":"active"},"limits":[{"id":"racl_001","tenant_id":"t_001","project_id":"proj_001","rule_kind":"subject_route","rule_id":"asub_001","currency":"USD","metric_scope":"tokens","period":"month","limit_amount":"1000","status":"active"}]}]}`))
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/projects/proj_001/usage-control-rules/limits/racl_001"):
			_, _ = w.Write([]byte(`{"deleted":true,"limit":{"id":"racl_001","tenant_id":"t_001","project_id":"proj_001","rule_kind":"subject_route","rule_id":"asub_001","currency":"USD","metric_scope":"tokens","period":"month","limit_amount":"1000","status":"disabled"}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects/proj_001/budget-leases"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"ble_001","run_id":"run_001","status":"active"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
		}
	}))
	defer server.Close()

	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
	ctx := context.Background()

	if _, err := client.Manager.CreateAccount(ctx, types.CreateAccountRequest{Name: "ACME"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.ListAccounts(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.BindSubject(ctx, types.BindAccountSubjectRequest{AccountID: "acct_001", SubjectType: "org", SubjectID: "org_001"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.ListAccountSubjects(ctx, "acct_001"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.CreateWallet(ctx, types.CreateWalletRequest{AccountID: "acct_001", Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.ListWallets(ctx, ListWalletsRequest{AccountID: "acct_001"}); err != nil {
		t.Fatal(err)
	}
	if snapshot, err := client.Manager.ReadVirtualWalletAmount(ctx, types.ReadVirtualWalletRequest{AccountID: "acct_001", Currency: "USD"}); err != nil {
		t.Fatal(err)
	} else if snapshot.AccountID != "acct_001" || snapshot.Balance.String() != "100" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if snapshot, err := client.Manager.ReadRealtimeWalletLimits(ctx, types.ReadVirtualWalletRequest{SubjectType: "org", SubjectID: "org_001", Currency: "USD"}); err != nil {
		t.Fatal(err)
	} else if snapshot.SubjectID != "org_001" || snapshot.Currency != "USD" || len(snapshot.Limits) != 1 || snapshot.Limits[0].MetricScope != "tokens" || snapshot.Limits[0].ID != "racl_001" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if snapshot, err := client.Manager.ReadVirtualWalletLimits(ctx, types.ReadVirtualWalletRequest{SubjectType: "org", SubjectID: "org_001", Currency: "USD"}); err != nil {
		t.Fatal(err)
	} else if snapshot.SubjectID != "org_001" || snapshot.Currency != "USD" || len(snapshot.Limits) != 1 || snapshot.Limits[0].MetricScope != "tokens" || snapshot.Limits[0].ID != "racl_001" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if snapshot, err := client.Manager.ReadSubjectRealtimeLimits(ctx, types.ReadSubjectRealtimeLimitsRequest{SubjectType: "org", SubjectID: "org_001"}); err != nil {
		t.Fatal(err)
	} else if len(snapshot.Limits) != 1 || snapshot.Limits[0].ID != "racl_001" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if _, err := client.Manager.ListWalletLedgerEntries(ctx, types.ListWalletLedgerEntriesRequest{AccountID: "acct_001", Limit: 20}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.CreditWallet(ctx, types.CreditWalletRequest{AccountID: "acct_001", Currency: "USD", Amount: decimal.RequireFromString("100")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.DebitWallet(ctx, types.DebitWalletRequest{AccountID: "acct_001", Currency: "USD", Amount: decimal.RequireFromString("10")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.SetWalletCreditLimit(ctx, types.SetWalletCreditLimitRequest{AccountID: "acct_001", Currency: "USD", CreditLimit: decimal.RequireFromString("50")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.CreateVoucherGrant(ctx, types.CreateVoucherGrantRequest{AccountID: "acct_001", Currency: "USD", Amount: decimal.RequireFromString("20")}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.ListVoucherGrants(ctx, ListVoucherGrantsRequest{AccountID: "acct_001", Currency: "USD", Limit: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.RevokeVoucherGrant(ctx, "vg_001"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.CreateUsageControlRule(ctx, types.CreateUsageControlRuleRequest{
		BillingAccountID: "acct_001",
		SubjectType:      "org",
		SubjectID:        "org_001",
		Currency:         "USD",
		Limits: []types.CreateRuleAccountLimitRequest{{
			MetricScope: "tokens",
			Period:      "month",
			LimitAmount: decimal.RequireFromString("1000"),
		}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.ListUsageControlRules(ctx, "acct_001"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Manager.DeleteRuleAccountLimit(ctx, "racl_001"); err != nil {
		t.Fatal(err)
	}
	if lease, err := client.Manager.CreateBudgetLease(ctx, types.CreateBudgetLeaseRequest{RunID: "run_001"}); err != nil || lease.ID != "ble_001" {
		t.Fatalf("lease=%#v err=%v", lease, err)
	}

	want := []string{
		"POST /v1/projects/proj_001/accounts",
		"GET /v1/projects/proj_001/accounts",
		"POST /v1/projects/proj_001/account-subjects",
		"GET /v1/projects/proj_001/account-subjects?account_id=acct_001&limit=500&offset=0",
		"POST /v1/projects/proj_001/wallets",
		"GET /v1/projects/proj_001/wallets?account_id=acct_001",
		"GET /v1/projects/proj_001/wallets/realtime/amount?account_id=acct_001&currency=USD",
		"GET /v1/projects/proj_001/subject/realtime/limits?subject_id=org_001&subject_type=org",
		"GET /v1/projects/proj_001/subject/realtime/limits?subject_id=org_001&subject_type=org",
		"GET /v1/projects/proj_001/subject/realtime/limits?subject_id=org_001&subject_type=org",
		"GET /v1/projects/proj_001/wallet-ledger-entries?account_id=acct_001&limit=20",
		"POST /v1/projects/proj_001/wallets/credit",
		"POST /v1/projects/proj_001/wallets/debit",
		"POST /v1/projects/proj_001/wallets/credit-limit",
		"POST /v1/projects/proj_001/voucher-grants",
		"GET /v1/projects/proj_001/voucher-grants?account_id=acct_001&currency=USD&limit=10",
		"POST /v1/projects/proj_001/voucher-grants/vg_001/revoke",
		"POST /v1/projects/proj_001/usage-control-rules",
		"GET /v1/projects/proj_001/usage-control-rules?account_id=acct_001&limit=500&offset=0",
		"DELETE /v1/projects/proj_001/usage-control-rules/limits/racl_001",
		"POST /v1/projects/proj_001/budget-leases",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v", requests)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q; all=%#v", i, requests[i], want[i], requests)
		}
	}
}

func TestIngestBuildsXBillingPayload(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/projects/proj_001/extractor-rule-sets/ers_001/events/ingest" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"raw_event_id":"raw_001","status":"received"}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "usage_key"})
	_, err := client.Ingest.Ingest(context.Background(), "ers_001", IngestEvent{
		Source:      "gateway",
		SubjectType: "organization",
		SubjectID:   "org_001",
		RawJSON: map[string]any{
			"meta": map[string]any{"org_id": "org_001"},
		},
		XBilling: &types.XBilling{
			Items: []types.ReportedChargeItem{{
				Metric:   "tool_call.web_search",
				Quantity: 2,
				Unit:     "call",
			}},
			PricingHints: map[string]types.ReportedPriceHint{
				"tool_call.web_search": {UnitPrice: 0.02, PricingUnit: 1, Currency: "USD"},
			},
		},
		PricingRuleSetID: "prs_fallback",
	})
	if err != nil {
		t.Fatal(err)
	}
	rawJSON := body["raw_json"].(map[string]any)
	if rawJSON["x-billing"] != nil {
		t.Fatalf("raw_json should not contain top-level x-billing request field: %#v", rawJSON)
	}
	if rawJSON["x-pricing-rule"] != "prs_fallback" {
		t.Fatalf("raw_json = %#v", rawJSON)
	}
	billing := body["x-billing"].(map[string]any)
	items := billing["items"].([]any)
	if items[0].(map[string]any)["metric"] != "tool_call.web_search" {
		t.Fatalf("x-billing = %#v", billing)
	}
	if body["subject_type"] != "organization" || body["subject_id"] != "org_001" {
		t.Fatalf("reported subject = %#v/%#v", body["subject_type"], body["subject_id"])
	}
}

func TestBatchIngestBuildsProjectPayload(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/projects/proj_001/events/batch-ingest" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"raw_event_ids":["raw_001","raw_002"],"count":2,"status":"received"}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "usage_key"})
	resp, err := client.Ingest.BatchIngest(context.Background(), "gateway", []BatchIngestEvent{
		{
			ExtractorRuleSetID: "ers_001",
			IngestEvent: IngestEvent{
				IdempotencyKey:   "req_001",
				SubjectType:      "organization",
				SubjectID:        "org_001",
				RawJSON:          map[string]any{"usage": map[string]any{"requests": 1}},
				PricingRuleSetID: "prs_fallback",
			},
		},
		{
			ExtractorRuleSetID: "ers_002",
			IngestEvent: IngestEvent{
				Source:         "gateway-override",
				IdempotencyKey: "req_002",
				RawJSON:        map[string]any{"usage": map[string]any{"requests": 2}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 2 || len(resp.RawEventIDs) != 2 {
		t.Fatalf("response = %#v", resp)
	}
	events := body["events"].([]any)
	first := events[0].(map[string]any)
	if first["extractor_rule_set_id"] != "ers_001" || first["idempotency_key"] != "req_001" || first["subject_type"] != "organization" || first["subject_id"] != "org_001" {
		t.Fatalf("first event = %#v", first)
	}
	rawJSON := first["raw_json"].(map[string]any)
	if rawJSON["x-pricing-rule"] != "prs_fallback" {
		t.Fatalf("raw_json = %#v", rawJSON)
	}
	if events[1].(map[string]any)["source"] != "gateway-override" {
		t.Fatalf("events = %#v", events)
	}
}

func TestBuildRawJSONSupportsLegacyEmbeddedXBillingPayload(t *testing.T) {
	raw, err := BuildRawJSON(map[string]any{
		"meta": map[string]any{"org_id": "org_001"},
	}, &types.XBilling{
		Items: []types.ReportedChargeItem{{
			Metric:   "tool_call.web_search",
			Quantity: 2,
			Unit:     "call",
		}},
		PricingHints: map[string]types.ReportedPriceHint{
			"tool_call.web_search": {UnitPrice: 0.02, PricingUnit: 1, Currency: "USD"},
		},
	}, "prs_fallback")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["x-pricing-rule"] != "prs_fallback" {
		t.Fatalf("payload = %#v", payload)
	}
	billing := payload["x-billing"].(map[string]any)
	items := billing["items"].([]any)
	if items[0].(map[string]any)["metric"] != "tool_call.web_search" {
		t.Fatalf("x-billing = %#v", billing)
	}
}

func TestQueryChartConvertsRowsToSeries(t *testing.T) {
	rows := []types.UsageAnalyticsRow{
		{Dimensions: map[string]string{"bucket": "60", "metric": "input_tokens"}, Measures: map[string]decimal.Decimal{"quantity_avg": decimal.RequireFromString("10")}},
		{Dimensions: map[string]string{"bucket": "0", "metric": "input_tokens"}, Measures: map[string]decimal.Decimal{"quantity_avg": decimal.RequireFromString("5")}},
	}
	series := chartSeriesFromRows(rows, "quantity_avg")
	if len(series) != 1 || series[0].Metric != "input_tokens" || len(series[0].Points) != 2 {
		t.Fatalf("series = %#v", series)
	}
	if series[0].Points[0].Time != 0 || !series[0].Points[0].Value.Equal(decimal.RequireFromString("5")) {
		t.Fatalf("points = %#v", series[0].Points)
	}
}

func TestQueryUsagePromotesMetricToMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/projects/proj_001/usage/query" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var req types.UsageAnalyticsQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Metric != "" || len(req.Metrics) != 1 || req.Metrics[0] != "input_tokens" {
			t.Fatalf("request = %#v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
	_, err := client.Query.Usage(context.Background(), types.UsageAnalyticsQueryRequest{
		Metric:    "input_tokens",
		StartTime: 1785000000,
		EndTime:   1785200000,
		Measures:  []string{"quantity"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryUsageRequiresMetricSelectorForMetricMeasures(t *testing.T) {
	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: "http://127.0.0.1:8080", APIKey: "service_key"})
	_, err := client.Query.Usage(context.Background(), types.UsageAnalyticsQueryRequest{
		StartTime: 1785000000,
		EndTime:   1785200000,
		Measures:  []string{"quantity"},
	})
	if err == nil || !strings.Contains(err.Error(), "metric or metrics is required for metric measures") {
		t.Fatalf("error = %v", err)
	}
}

func TestQueryBillRequiresMetrics(t *testing.T) {
	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: "http://127.0.0.1:8080", APIKey: "service_key"})
	_, err := client.Query.Bill(context.Background(), types.UsageBillQueryRequest{
		StartTime: 1785000000,
		EndTime:   1785200000,
	})
	if err == nil || !strings.Contains(err.Error(), "metrics are required for bill queries") {
		t.Fatalf("error = %v", err)
	}
}

func TestQueryBillUsesDedicatedBillRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/projects/proj_001/usage/bills/query" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var req types.UsageBillQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.BucketSize != "5m" || req.Timezone != "Asia/Shanghai" || req.BillingAccountID != "acct_001" || len(req.Metrics) != 1 {
			t.Fatalf("request = %#v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[{"dimensions":{"day":"2026-07-27"},"request_count":2,"metrics":{"input_tokens":{"quantity":"12","amount":"0.01"}},"amount":"0.01"}],"has_more":false}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
	resp, err := client.Query.Bill(context.Background(), types.UsageBillQueryRequest{
		BucketSize:       "5m",
		Timezone:         "Asia/Shanghai",
		BillingAccountID: "acct_001",
		Metrics:          []string{"input_tokens"},
		StartTime:        1785000000,
		EndTime:          1785200000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Rows) != 1 || resp.Rows[0].RequestCount != 2 || !resp.Rows[0].Amount.Equal(decimal.RequireFromString("0.01")) {
		t.Fatalf("response = %#v", resp)
	}
}
