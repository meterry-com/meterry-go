package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sdk "github.com/meterry-com/meterry-go"
	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestQueryAggregatePreservesRequestAndDecodesCursorPages(t *testing.T) {
	const cursor = "opaque+/=:cursor"
	const dimensions = `"dimensions":{"bucket":"1785081600","subject_type":"user","subject_id":"user_123","model":"m","provider":"p","region":"r","team":"t","service":"s"}`
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.RequestURI() != "/gateway/v1/projects/proj_001/usage/aggregate/query" {
			t.Errorf("request = %s %s", r.Method, r.URL.RequestURI())
		}
		if r.Header.Get("Authorization") != "Bearer service_key" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("request headers = %#v", r.Header)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- body
		w.Header().Set("Content-Type", "application/json")
		if _, hasCursor := body["cursor"]; !hasCursor {
			fmt.Fprint(w, `{"rows":[{`+dimensions+`,"request_count":2,"analytics_only":false,"metrics":{"metric_0":{"quantity":"0","amount":"0","quantity_avg":"0","amount_avg":"0","item_count":1,"analytics_only":false},"cost.tool_call":{"quantity":"1.234567890123456789","amount":"0.000000000000000001","quantity_avg":"0.6172839450617283945","amount_avg":"0.0000000000000000005","item_count":2,"analytics_only":false}},"amount":"0.000000000000000001","analytics_only_amount":"0"}],"has_more":true,"next_cursor":"`+cursor+`"}`)
		} else {
			fmt.Fprint(w, `{"rows":[{`+dimensions+`,"request_count":2,"analytics_only":true,"metrics":{"cost.tool_call":{"amount":"1","amount_avg":"0.5","item_count":2,"analytics_only":true,"unavailable_measures":["quantity","quantity_avg"]}},"amount":"0","analytics_only_amount":"1"}],"has_more":false}`)
		}
	}))
	defer server.Close()
	client := sdk.MustNewClient(sdk.Config{ProjectID: "proj_001", BaseURL: server.URL + "/gateway", APIKey: "service_key"})
	metrics := make([]string, 50)
	for i := range metrics {
		metrics[i] = fmt.Sprintf("metric_%d", i)
	}
	metrics[49] = "cost.tool_call"
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		t.Fatal(err)
	}
	// Assert the wire contract independently of the SDK request struct's tags.
	var expectedRequest map[string]any
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"start_time":1785081600,"end_time":1785686400,"metrics":%s,"measures":["quantity","amount","quantity_avg","amount_avg","item_count","request_count"],"billing_account_id":"acct_001","subject_type":"user","subject_id":"user_123","filters":{"provider":"p"},"group_by":["bucket","subject_type","subject_id","model","provider","region","team","service"],"bucket_size":"day","time_order":"desc","timezone":"Asia/Shanghai","limit":1}`, metricsJSON)), &expectedRequest); err != nil {
		t.Fatal(err)
	}
	req := types.UsageAggregateQueryRequest{
		StartTime: 1785081600, EndTime: 1785686400,
		Metrics: metrics, Measures: []string{"quantity", "amount", "quantity_avg", "amount_avg", "item_count", "request_count"},
		BillingAccountID: "acct_001", SubjectType: "user", SubjectID: "user_123", Filters: map[string]string{"provider": "p"},
		GroupBy:    []string{"bucket", "subject_type", "subject_id", "model", "provider", "region", "team", "service"},
		BucketSize: "day", TimeOrder: "desc", Timezone: "Asia/Shanghai", Limit: 1,
	}
	for page := 0; page < 2; page++ {
		resp, err := client.Query.Aggregate(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		select {
		case got := <-requests:
			if !reflect.DeepEqual(got, expectedRequest) {
				t.Fatalf("request JSON = %#v, want %#v", got, expectedRequest)
			}
		default:
			t.Fatal("aggregate request did not reach the server")
		}
		if len(resp.Rows) != 1 {
			t.Fatalf("response = %#v", resp)
		}
		row := resp.Rows[0]
		if row.AnalyticsOnly == nil || *row.AnalyticsOnly != (page == 1) || row.RequestCount != 2 || len(row.Dimensions) != 8 || row.Dimensions["service"] != "s" {
			t.Fatalf("row = %#v", row)
		}
		metric, ok := row.Metrics["cost.tool_call"]
		if !ok || metric.AnalyticsOnly != (page == 1) || metric.ItemCount == nil || *metric.ItemCount != 2 {
			t.Fatalf("metric = %#v", metric)
		}
		if page == 0 {
			if !resp.HasMore || resp.NextCursor != cursor || row.Amount == nil || row.Amount.String() != "0.000000000000000001" || row.AnalyticsOnlyAmount == nil || !row.AnalyticsOnlyAmount.IsZero() {
				t.Fatalf("first page = %#v", resp)
			}
			if metric.Quantity == nil || metric.Quantity.String() != "1.234567890123456789" || metric.Amount == nil || metric.Amount.String() != "0.000000000000000001" || metric.QuantityAvg == nil || metric.QuantityAvg.String() != "0.6172839450617283945" || metric.AmountAvg == nil || metric.AmountAvg.String() != "0.0000000000000000005" {
				t.Fatalf("decimal measures = %#v", metric)
			}
			zero := row.Metrics["metric_0"]
			if zero.Quantity == nil || !zero.Quantity.IsZero() || zero.Amount == nil || !zero.Amount.IsZero() || zero.QuantityAvg == nil || !zero.QuantityAvg.IsZero() || zero.AmountAvg == nil || !zero.AmountAvg.IsZero() {
				t.Fatalf("explicit zero measures = %#v", zero)
			}
		} else {
			if resp.HasMore || resp.NextCursor != "" || row.Amount == nil || !row.Amount.IsZero() || row.AnalyticsOnlyAmount == nil || row.AnalyticsOnlyAmount.String() != "1" {
				t.Fatalf("last page = %#v", resp)
			}
			if metric.Quantity != nil || metric.QuantityAvg != nil || metric.Amount == nil || metric.Amount.String() != "1" || !reflect.DeepEqual(metric.UnavailableMeasures, []string{"quantity", "quantity_avg"}) {
				t.Fatalf("unavailable measures = %#v", metric)
			}
		}
		req.Cursor = resp.NextCursor
		expectedRequest["cursor"] = cursor
	}
}

func TestQueryAggregateRequestCountWithoutMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if _, ok := body["metrics"]; ok {
			t.Errorf("metrics must be omitted: %s", body["metrics"])
		}
		if string(body["measures"]) != `["request_count"]` {
			t.Errorf("measures = %s", body["measures"])
		}
		fmt.Fprint(w, `{"rows":[{"dimensions":{},"request_count":0,"metrics":{}}],"has_more":false}`)
	}))
	defer server.Close()
	client := sdk.MustNewClient(sdk.Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
	resp, err := client.Query.Aggregate(context.Background(), types.UsageAggregateQueryRequest{
		StartTime: 1785081600, EndTime: 1785686400, Measures: []string{"request_count"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Rows) != 1 || resp.HasMore || resp.NextCursor != "" {
		t.Fatalf("response = %#v", resp)
	}
	row := resp.Rows[0]
	if row.RequestCount != 0 || row.AnalyticsOnly != nil || row.Amount != nil || row.AnalyticsOnlyAmount != nil || len(row.Metrics) != 0 {
		t.Fatalf("request-count row = %#v", row)
	}
}

func TestQueryAggregatePropagatesAPIErrors(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				fmt.Fprint(w, `{"error":"aggregate query failed"}`)
			}))
			defer server.Close()
			client := sdk.MustNewClient(sdk.Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
			_, err := client.Query.Aggregate(context.Background(), types.UsageAggregateQueryRequest{
				StartTime: 1785081600, EndTime: 1785686400, Metrics: []string{"input_tokens"},
			})
			var apiErr *sdk.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status || apiErr.Message != "aggregate query failed" {
				t.Fatalf("error = %v, want API error %d", err, status)
			}
		})
	}
}

func TestQueryAggregatePropagatesCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("cancelled request reached the server")
	}))
	defer server.Close()
	client := sdk.MustNewClient(sdk.Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "service_key"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Query.Aggregate(ctx, types.UsageAggregateQueryRequest{
		StartTime: 1785081600, EndTime: 1785686400, Metrics: []string{"input_tokens"},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}
