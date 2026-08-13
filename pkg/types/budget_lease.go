package types

import "github.com/shopspring/decimal"

// BudgetLease is a one-time, per-run amount budget. It is deliberately
// separate from AccountSubject: the latter remains a durable attribution and
// payer route, while a lease is short lived control state for one agent run.
type BudgetLease struct {
	ID               string            `json:"id"`
	TenantID         string            `json:"tenant_id"`
	ProjectID        string            `json:"project_id"`
	RunID            string            `json:"run_id"`
	BillingAccountID string            `json:"billing_account_id"`
	SubjectType      string            `json:"subject_type"`
	SubjectID        string            `json:"subject_id"`
	Currency         string            `json:"currency"`
	LimitAmount      decimal.Decimal   `json:"limit_amount"`
	Action           string            `json:"action"`
	WebhookEvent     string            `json:"webhook_event,omitempty"`
	Status           BudgetLeaseStatus `json:"status"`
	CreatedAt        int64             `json:"created_at"`
	UpdatedAt        int64             `json:"updated_at"`
}

type BudgetLeaseStatus string

const (
	BudgetLeaseStatusActive    BudgetLeaseStatus = "active"
	BudgetLeaseStatusBlocked   BudgetLeaseStatus = "blocked"
	BudgetLeaseStatusCompleted BudgetLeaseStatus = "completed"
	BudgetLeaseStatusSettling  BudgetLeaseStatus = "settling"
	BudgetLeaseStatusArchived  BudgetLeaseStatus = "archived"
)

type CreateBudgetLeaseRequest struct {
	ID               string          `json:"id,omitempty"`
	RunID            string          `json:"run_id"`
	BillingAccountID string          `json:"billing_account_id"`
	SubjectType      string          `json:"subject_type"`
	SubjectID        string          `json:"subject_id"`
	Currency         string          `json:"currency"`
	LimitAmount      decimal.Decimal `json:"limit_amount"`
	Action           string          `json:"action,omitempty"`
	WebhookEvent     string          `json:"webhook_event,omitempty"`
}

type CreateBudgetLeaseCommand struct {
	ID               string
	TenantID         string
	ProjectID        string
	RunID            string
	BillingAccountID string
	SubjectType      string
	SubjectID        string
	Currency         string
	LimitAmount      decimal.Decimal
	Action           string
	WebhookEvent     string
}

// XUsageControl is attached to an ingest event. It is embedded in RawJSON at
// ingest time so queued delivery, replay, and audit retain the lease binding.
type XUsageControl struct {
	BudgetLeaseID string `json:"budget_lease_id,omitempty"`
}
