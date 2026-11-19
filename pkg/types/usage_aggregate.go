package types

import "github.com/shopspring/decimal"

// UsageAggregateQueryRequest queries the usage/aggregate/query endpoint.
// GroupBy supports up to eight distinct dimensions. Metrics supports up to 50
// distinct names and may be omitted only for request_count-only queries.
// Fields, types, and JSON tags match the API request.
type UsageAggregateQueryRequest struct {
	StartTime        int64             `json:"start_time"`
	EndTime          int64             `json:"end_time"`
	Metrics          []string          `json:"metrics,omitempty"`
	Measures         []string          `json:"measures,omitempty"`
	BillingAccountID string            `json:"billing_account_id,omitempty"`
	SubjectType      string            `json:"subject_type,omitempty"`
	SubjectID        string            `json:"subject_id,omitempty"`
	Filters          map[string]string `json:"filters,omitempty"`
	GroupBy          []string          `json:"group_by,omitempty"`
	BucketSize       string            `json:"bucket_size,omitempty"`
	TimeOrder        string            `json:"time_order,omitempty"`
	Timezone         string            `json:"timezone,omitempty"`
	Cursor           string            `json:"cursor,omitempty"`
	Limit            int               `json:"limit,omitempty"`
}

// UsageAggregateValues contains only requested measures. Historical mixed
// metrics without quantity subtotals explicitly identify unavailable measures.
type UsageAggregateValues struct {
	Quantity            *decimal.Decimal `json:"quantity,omitempty"`
	Amount              *decimal.Decimal `json:"amount,omitempty"`
	QuantityAvg         *decimal.Decimal `json:"quantity_avg,omitempty"`
	AmountAvg           *decimal.Decimal `json:"amount_avg,omitempty"`
	ItemCount           *uint64          `json:"item_count,omitempty"`
	UnavailableMeasures []string         `json:"unavailable_measures,omitempty"`
}

type UsageAggregateMetricResult struct {
	UsageAggregateValues
	AnalyticsOnly bool `json:"analytics_only"`
}

type UsageAggregateResultRow struct {
	// AnalyticsOnly identifies an automatically classified group. It is absent
	// only for request-count queries that do not select metrics.
	AnalyticsOnly       *bool                                 `json:"analytics_only,omitempty"`
	Dimensions          map[string]string                     `json:"dimensions"`
	RequestCount        uint64                                `json:"request_count"`
	Metrics             map[string]UsageAggregateMetricResult `json:"metrics"`
	Amount              *decimal.Decimal                      `json:"amount,omitempty"`
	AnalyticsOnlyAmount *decimal.Decimal                      `json:"analytics_only_amount,omitempty"`
}

type UsageAggregateQueryResponse struct {
	Rows       []UsageAggregateResultRow `json:"rows"`
	NextCursor string                    `json:"next_cursor,omitempty"`
	HasMore    bool                      `json:"has_more"`
}
