package types

import "github.com/shopspring/decimal"

// UsageLimitWebhookPayload decodes both exhausted and recovered notifications.
// Subject identifies the member; Limit's dimension identifies an optional key.
type UsageLimitWebhookPayload struct {
	ID        string           `json:"id,omitempty"`
	EventType WebhookEventType `json:"event_type"`
	TenantID  string           `json:"tenant_id"`
	ProjectID string           `json:"project_id"`
	Account   struct {
		ID string `json:"id"`
	} `json:"account"`
	Subject struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"subject"`
	Limit    UsageLimitWebhookLimit `json:"limit"`
	Recovery *struct {
		Reason      string `json:"reason"`
		RecoveredAt int64  `json:"recovered_at"`
	} `json:"recovery,omitempty"`
}

type UsageLimitWebhookLimit struct {
	ID                 string             `json:"id"`
	RuleKind           UsageLimitRuleKind `json:"rule_kind,omitempty"`
	DimensionKey       string             `json:"dimension_key,omitempty"`
	DimensionValue     string             `json:"dimension_value,omitempty"`
	Measure            string             `json:"measure"`
	SelectorType       string             `json:"selector_type"`
	SelectorValue      string             `json:"selector_value,omitempty"`
	Currency           string             `json:"currency"`
	Period             string             `json:"period"`
	WindowID           string             `json:"window_id"`
	UsedAmount         decimal.Decimal    `json:"used_amount"`
	PreviousUsedAmount decimal.Decimal    `json:"previous_used_amount"`
	LimitAmount        decimal.Decimal    `json:"limit_amount"`
	RemainingAmount    decimal.Decimal    `json:"remaining_amount"`
	Action             string             `json:"action"`
	NextResetAt        int64              `json:"next_reset_at,omitempty"`
}
