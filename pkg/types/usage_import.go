package types

import "github.com/shopspring/decimal"

type ImportSubjectUsageRequest struct {
	LimitID    string          `json:"limit_id"`
	UsedAmount decimal.Decimal `json:"used_amount"`
	// CoveredBefore is exclusive business time: the baseline includes all usage before it.
	CoveredBefore    int64  `json:"covered_before"`
	WindowStart      int64  `json:"window_start,omitempty"`
	WindowEnd        int64  `json:"window_end,omitempty"`
	ExpectedImportID string `json:"expected_import_id,omitempty"`
	IdempotencyKey   string `json:"idempotency_key"`
}

type SubjectUsageImport struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	ProjectID string `json:"project_id"`
	ImportSubjectUsageRequest
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	AppliedAt int64  `json:"applied_at,omitempty"`
}

type UpdateUsageControlLimitRequest struct {
	LimitAmount  *decimal.Decimal `json:"limit_amount,omitempty"`
	Period       *string          `json:"period,omitempty"`
	PeriodType   *string          `json:"period_type,omitempty"`
	Timezone     *string          `json:"timezone,omitempty"`
	Action       *string          `json:"action,omitempty"`
	WebhookEvent *string          `json:"webhook_event,omitempty"`
}
