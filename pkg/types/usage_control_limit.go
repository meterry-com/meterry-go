package types

import "github.com/shopspring/decimal"

type UpdateUsageControlLimitRequest struct {
	LimitAmount  *decimal.Decimal `json:"limit_amount,omitempty"`
	Period       *string          `json:"period,omitempty"`
	PeriodType   *string          `json:"period_type,omitempty"`
	Timezone     *string          `json:"timezone,omitempty"`
	Action       *string          `json:"action,omitempty"`
	WebhookEvent *string          `json:"webhook_event,omitempty"`
}
