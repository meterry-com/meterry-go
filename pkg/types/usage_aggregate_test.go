package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestUsageAggregateQueryRequestJSONContract(t *testing.T) {
	for _, test := range []struct {
		name string
		req  types.UsageAggregateQueryRequest
		want string
	}{
		{
			name: "all fields",
			req: types.UsageAggregateQueryRequest{
				StartTime: 1785081600, EndTime: 1785686400,
				Metrics: []string{"input_tokens", "cost.tool_call"}, Measures: []string{"quantity", "amount", "request_count"},
				BillingAccountID: "acct_001", SubjectType: "user", SubjectID: "user_123",
				Filters: map[string]string{"provider": "openai"}, GroupBy: []string{"bucket", "model"},
				BucketSize: "day", TimeOrder: "desc", Timezone: "Asia/Shanghai", Cursor: "opaque+/=:cursor", Limit: 100,
			},
			want: `{"start_time":1785081600,"end_time":1785686400,"metrics":["input_tokens","cost.tool_call"],"measures":["quantity","amount","request_count"],"billing_account_id":"acct_001","subject_type":"user","subject_id":"user_123","filters":{"provider":"openai"},"group_by":["bucket","model"],"bucket_size":"day","time_order":"desc","timezone":"Asia/Shanghai","cursor":"opaque+/=:cursor","limit":100}`,
		},
		{
			name: "request counts without metric selectors",
			req: types.UsageAggregateQueryRequest{
				StartTime: 1785081600, EndTime: 1785686400, Measures: []string{"request_count"},
			},
			want: `{"start_time":1785081600,"end_time":1785686400,"measures":["request_count"]}`,
		},
		{
			name: "empty optionals are omitted and timestamps remain required",
			req:  types.UsageAggregateQueryRequest{Metrics: []string{}, Measures: []string{}, GroupBy: []string{}, Filters: map[string]string{}},
			want: `{"start_time":0,"end_time":0}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.req)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("request JSON = %s, want %s", body, test.want)
			}
		})
	}
}
