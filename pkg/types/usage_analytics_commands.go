package types

import "github.com/shopspring/decimal"

type UsageAnalyticsQueryCommand struct {
	TenantID    string
	ProjectID   string
	ReportID    string
	BucketSize  string
	SubjectType string
	SubjectID   string
	Metric      string
	Metrics     []string
	StartTime   int64
	EndTime     int64
	Filters     map[string]string
	GroupBy     []string
	Measures    []string
	Limit       int
}

type UsageBillQueryCommand struct {
	TenantID         string
	ProjectID        string
	BucketSize       string
	Timezone         string
	Cursor           string
	BillingAccountID string
	SubjectType      string
	SubjectID        string
	Metrics          []string
	StartTime        int64
	EndTime          int64
	Filters          map[string]string
	GroupBy          []string
	Limit            int
}

type UsageAggregateQuery struct {
	TenantID               string
	ProjectID              string
	BucketSize             string
	Timezone               string
	BillingAccountID       string
	SubjectType            string
	SubjectID              string
	Metrics                []string
	StartTime              int64
	EndTime                int64
	Filters                map[string]string
	GroupBy                []string
	After                  []string
	Limit                  int
	IncludeMetrics         bool
	IncludeMetricCount     bool
	IncludeRequestCount    bool
	IncludeUsageEventCount bool
}

type UsageAggregateRow struct {
	Dimensions      map[string]string
	RequestCount    uint64
	UsageEventCount uint64
	Metrics         map[string]UsageAggregateMetric
}

type UsageAggregateMetric struct {
	Quantity decimal.Decimal
	Amount   decimal.Decimal
	Count    uint64
}
