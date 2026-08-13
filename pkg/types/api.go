package types

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type IngestEventRequest struct {
	Source          string            `json:"source"`
	ExternalEventID string            `json:"external_event_id,omitempty"`
	IdempotencyKey  string            `json:"idempotency_key,omitempty"`
	SubjectType     string            `json:"subject_type,omitempty"`
	SubjectID       string            `json:"subject_id,omitempty"`
	OccurredAt      int64             `json:"occurred_at"`
	RawJSON         json.RawMessage   `json:"raw_json"`
	XBilling        *XBilling         `json:"x-billing,omitempty"`
	XUsageControl   *XUsageControl    `json:"x-usage-control,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type IngestEventResponse struct {
	RawEventID string         `json:"raw_event_id"`
	Status     RawEventStatus `json:"status"`
}

type BatchIngestEventRequest struct {
	ExtractorRuleSetID string            `json:"extractor_rule_set_id"`
	Source             string            `json:"source,omitempty"`
	ExternalEventID    string            `json:"external_event_id,omitempty"`
	IdempotencyKey     string            `json:"idempotency_key"`
	SubjectType        string            `json:"subject_type,omitempty"`
	SubjectID          string            `json:"subject_id,omitempty"`
	OccurredAt         int64             `json:"occurred_at"`
	RawJSON            json.RawMessage   `json:"raw_json"`
	XBilling           *XBilling         `json:"x-billing,omitempty"`
	XUsageControl      *XUsageControl    `json:"x-usage-control,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

type BatchIngestEventsRequest struct {
	Source string                    `json:"source,omitempty"`
	Events []BatchIngestEventRequest `json:"events"`
}

type BatchIngestEventsResponse struct {
	RawEventIDs []string       `json:"raw_event_ids"`
	Count       int            `json:"count"`
	Status      RawEventStatus `json:"status"`
}

type TestRulesRequest struct {
	ExtractorRuleSetID string          `json:"extractor_rule_set_id,omitempty"`
	PricingRuleSetID   string          `json:"pricing_rule_set_id,omitempty"`
	RawJSON            json.RawMessage `json:"raw_json"`
}

type CreateExtractorRuleSetRequest struct {
	ID            string                `json:"id,omitempty"`
	Name          string                `json:"name,omitempty"`
	TemplateID    string                `json:"template_id,omitempty"`
	Version       int                   `json:"version,omitempty"`
	Status        RuleSetStatus         `json:"status,omitempty"`
	EffectiveFrom int64                 `json:"effective_from,omitempty"`
	Subject       SubjectExtractor      `json:"subject,omitempty"`
	Dimensions    []DimensionExtractor  `json:"dimensions,omitempty"`
	ChargeItems   []ChargeItemExtractor `json:"charge_item_extractors,omitempty"`
}

type CreateExtractorRuleSetTemplateRequest struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Fields      json.RawMessage `json:"fields"`
}

type UpdateExtractorRuleSetTemplateRequest struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Fields      json.RawMessage `json:"fields,omitempty"`
}

type ListExtractorRuleSetTemplatesResponse struct {
	Templates []ExtractorRuleSetTemplate `json:"templates"`
}

type ListExtractorRuleSetsResponse struct {
	RuleSets []ExtractorRuleSet `json:"rule_sets"`
}

type CreatePricingRuleSetRequest struct {
	ExtractorRuleSetID string                     `json:"extractor_rule_set_id"`
	Name               string                     `json:"name,omitempty"`
	Currency           string                     `json:"currency,omitempty"`
	EffectiveFrom      int64                      `json:"effective_from,omitempty"`
	Rules              []CreatePricingRuleRequest `json:"rules,omitempty"`
}

type CreatePricingRuleRequest struct {
	Name       string `json:"name"`
	Priority   int    `json:"priority"`
	Metric     string `json:"metric,omitempty"`
	MatchExpr  string `json:"match_expr"`
	AmountExpr string `json:"amount_expr"`
	UnitPrice  string `json:"unit_price_expr,omitempty"`
}

type ListPricingRuleSetsResponse struct {
	RuleSets []PricingRuleSet `json:"rule_sets"`
}

type DeleteRuleSetResponse struct {
	Deleted  bool          `json:"deleted"`
	Archived bool          `json:"archived"`
	Status   RuleSetStatus `json:"status,omitempty"`
}

type TestRulesResponse struct {
	SubjectType    string            `json:"subject_type,omitempty"`
	SubjectID      string            `json:"subject_id,omitempty"`
	Dimensions     map[string]string `json:"dimensions"`
	ChargeItems    []ChargeItem      `json:"charge_items"`
	UnmatchedItems []ChargeItem      `json:"unmatched_items"`
}

type ProcessRawEventResponse struct {
	RawEventID     string            `json:"raw_event_id"`
	UsageEventID   string            `json:"usage_event_id"`
	SubjectType    string            `json:"subject_type,omitempty"`
	SubjectID      string            `json:"subject_id,omitempty"`
	Dimensions     map[string]string `json:"dimensions"`
	ChargeItems    []ChargeItem      `json:"charge_items"`
	UnmatchedItems []ChargeItem      `json:"unmatched_items"`
}

type ProcessRawEventsBatchResponse struct {
	Processed                int      `json:"processed"`
	Failed                   int      `json:"failed"`
	UsageEventIDs            []string `json:"usage_event_ids"`
	UsageVersions            []int64  `json:"usage_versions,omitempty"`
	RatedItemCount           int      `json:"rated_item_count"`
	ChargeItemCount          int      `json:"charge_item_count"`
	AnalyticsMessageCount    int      `json:"analytics_message_count,omitempty"`
	AnalyticsPublishFailures int      `json:"analytics_publish_failures,omitempty"`
	VirtualDebitCount        int      `json:"virtual_debit_count,omitempty"`
	VirtualDebitFailures     int      `json:"virtual_debit_failures,omitempty"`
	UnmatchedMetrics         []string `json:"unmatched_metrics,omitempty"`
}

type ReconcileUsageChargesResponse struct {
	StartVersion       int64    `json:"start_version"`
	EndVersion         int64    `json:"end_version"`
	InterimBills       int      `json:"interim_bills"`
	AppliedWallets     int      `json:"applied_wallets"`
	SkippedWallets     int      `json:"skipped_wallets"`
	AppliedVouchers    int      `json:"applied_vouchers,omitempty"`
	SkippedVouchers    int      `json:"skipped_vouchers,omitempty"`
	LedgerEntries      int      `json:"ledger_entries"`
	RedisResetCount    int      `json:"redis_reset_count,omitempty"`
	RedisResetFailures int      `json:"redis_reset_failures,omitempty"`
	UnresolvedSubjects []string `json:"unresolved_subjects,omitempty"`
}

type ListUsageLogsRequest struct {
	TenantID     string
	ProjectID    string
	UsageEventID string
	SubjectType  string
	SubjectID    string
	Metric       string
	Metrics      []string
	Filters      map[string]string
	StartTime    int64
	EndTime      int64
	Limit        int
}

type ListUsageLogsResponse struct {
	UsageLogs []UsageLogRow `json:"usage_logs"`
}

type RepublishUsageLogMessagesRequest struct {
	RawEventIDs []string `json:"raw_event_ids"`
	Limit       int      `json:"limit,omitempty"`
}

type RepublishUsageLogMessagesCommand struct {
	TenantID    string
	ProjectID   string
	RawEventIDs []string
	Limit       int
}

type RepublishUsageLogMessagesResponse struct {
	OutboxCount     int `json:"outbox_count"`
	MessageCount    int `json:"message_count"`
	PublishFailures int `json:"publish_failures,omitempty"`
}

type ListUsageEventLogsRequest struct {
	TenantID     string
	ProjectID    string
	UsageEventID string
	RawEventID   string
	SubjectType  string
	SubjectID    string
	Metric       string
	Metrics      []string
	Filters      map[string]string
	StartTime    int64
	EndTime      int64
	AfterID      string
	Limit        int
}

type ListUsageEventLogsResponse struct {
	UsageEvents []UsageEventLogRow `json:"usage_events"`
}

// UsageEventDetail is a usage event log row returned directly at the root of a
// detail response. Its metrics are the complete, aggregated charge-item view.
type UsageEventDetail UsageEventLogRow

type ListUsageEventDetailsResponse struct {
	UsageEvents []UsageEventDetail `json:"usage_events"`
	NextAfterID string             `json:"next_after_id,omitempty"`
}

type UsageAnalyticsQueryRequest struct {
	ReportID    string            `json:"report_id,omitempty"`
	BucketSize  string            `json:"bucket_size,omitempty"`
	SubjectType string            `json:"subject_type,omitempty"`
	SubjectID   string            `json:"subject_id,omitempty"`
	Metric      string            `json:"metric,omitempty"`
	Metrics     []string          `json:"metrics,omitempty"`
	StartTime   int64             `json:"start_time,omitempty"`
	EndTime     int64             `json:"end_time,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Measures    []string          `json:"measures,omitempty"`
	Limit       int               `json:"limit,omitempty"`
}

type UsageBillQueryRequest struct {
	BucketSize       string            `json:"bucket_size,omitempty"`
	Timezone         string            `json:"timezone,omitempty"`
	Cursor           string            `json:"cursor,omitempty"`
	BillingAccountID string            `json:"billing_account_id,omitempty"`
	SubjectType      string            `json:"subject_type,omitempty"`
	SubjectID        string            `json:"subject_id,omitempty"`
	Metrics          []string          `json:"metrics,omitempty"`
	StartTime        int64             `json:"start_time,omitempty"`
	EndTime          int64             `json:"end_time,omitempty"`
	Filters          map[string]string `json:"filters,omitempty"`
	GroupBy          []string          `json:"group_by,omitempty"`
	Limit            int               `json:"limit,omitempty"`
}

type UsageAnalyticsRow struct {
	Dimensions map[string]string          `json:"dimensions"`
	Measures   map[string]decimal.Decimal `json:"measures"`
}

type UsageAnalyticsQueryResponse struct {
	Rows []UsageAnalyticsRow `json:"rows"`
}

type UsageBillMetric struct {
	Quantity decimal.Decimal `json:"quantity"`
	Amount   decimal.Decimal `json:"amount"`
}

type UsageBillRow struct {
	Dimensions   map[string]string          `json:"dimensions"`
	RequestCount uint64                     `json:"request_count"`
	Metrics      map[string]UsageBillMetric `json:"metrics"`
	Amount       decimal.Decimal            `json:"amount"`
}

type UsageBillQueryResponse struct {
	Rows       []UsageBillRow `json:"rows"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type Project struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	ProjectID   string        `json:"project_id"`
	Name        string        `json:"name"`
	Environment string        `json:"environment"`
	Description string        `json:"description,omitempty"`
	Status      RuleSetStatus `json:"status"`
	CreatedAt   int64         `json:"created_at"`
	UpdatedAt   int64         `json:"updated_at"`
}

type CreateProjectRequest struct {
	ID          string        `json:"id,omitempty"`
	ProjectID   string        `json:"project_id,omitempty"`
	Name        string        `json:"name"`
	Environment string        `json:"environment,omitempty"`
	Description string        `json:"description,omitempty"`
	Status      RuleSetStatus `json:"status,omitempty"`
}

type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type UsageReport struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	ProjectID   string            `json:"project_id"`
	ReportID    string            `json:"report_id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	SubjectType string            `json:"subject_type,omitempty"`
	BucketSize  string            `json:"bucket_size"`
	Filters     map[string]string `json:"filters,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Measures    []string          `json:"measures,omitempty"`
	Metrics     []string          `json:"metrics,omitempty"`
	Status      RuleSetStatus     `json:"status"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

type CreateUsageReportRequest struct {
	ID          string            `json:"id,omitempty"`
	ReportID    string            `json:"report_id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	SubjectType string            `json:"subject_type,omitempty"`
	BucketSize  string            `json:"bucket_size,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Measures    []string          `json:"measures,omitempty"`
	Metrics     []string          `json:"metrics,omitempty"`
	Status      RuleSetStatus     `json:"status,omitempty"`
}
