package types

import "github.com/shopspring/decimal"

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusDisabled AccountStatus = "disabled"
)

type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "active"
	WalletStatusDisabled WalletStatus = "disabled"
)

type LedgerDirection string

const (
	LedgerDirectionDebit  LedgerDirection = "debit"
	LedgerDirectionCredit LedgerDirection = "credit"
)

type FundType string

const (
	FundTypeCash    FundType = "cash"
	FundTypeVoucher FundType = "voucher"
)

type VoucherGrantStatus string

const (
	VoucherGrantStatusActive  VoucherGrantStatus = "active"
	VoucherGrantStatusRevoked VoucherGrantStatus = "revoked"
	VoucherGrantStatusExpired VoucherGrantStatus = "expired"
)

type Account struct {
	ID                 string        `json:"id"`
	TenantID           string        `json:"tenant_id"`
	ProjectID          string        `json:"project_id"`
	Name               string        `json:"name,omitempty"`
	PrimarySubjectType string        `json:"primary_subject_type,omitempty"`
	PrimarySubjectID   string        `json:"primary_subject_id,omitempty"`
	Status             AccountStatus `json:"status"`
	CreatedAt          int64         `json:"created_at"`
	UpdatedAt          int64         `json:"updated_at"`
}

type AccountSubject struct {
	ID                 string           `json:"id"`
	TenantID           string           `json:"tenant_id"`
	ProjectID          string           `json:"project_id"`
	AccountID          string           `json:"account_id"`
	SubjectType        string           `json:"subject_type"`
	SubjectID          string           `json:"subject_id"`
	EffectiveFrom      int64            `json:"effective_from"`
	EffectiveTo        int64            `json:"effective_to,omitempty"`
	CreatedAt          int64            `json:"created_at"`
	UsageKind          UsageControlKind `json:"usage_kind"`
	UsageID            string           `json:"usage_id,omitempty"`
	RouteConfigVersion int              `json:"route_config_version"`
	// AccountUsageLimitReset* is copied to every active route when its billing
	// account is reset. SubjectUsageLimitReset* is the route-only boundary.
	// Keeping both snapshots on the route makes them part of the normal route
	// cache read instead of an additional realtime Redis lookup.
	AccountUsageLimitResetGeneration          int64 `json:"account_usage_limit_reset_generation,omitempty"`
	AccountUsageLimitResetEffectiveOccurredAt int64 `json:"account_usage_limit_reset_effective_occurred_at,omitempty"`
	SubjectUsageLimitResetGeneration          int64 `json:"subject_usage_limit_reset_generation,omitempty"`
	SubjectUsageLimitResetEffectiveOccurredAt int64 `json:"subject_usage_limit_reset_effective_occurred_at,omitempty"`
}

type UsageControlKind string

const (
	UsageControlKindNone   UsageControlKind = "none"
	UsageControlKindDirect UsageControlKind = "direct"
	UsageControlKindPlan   UsageControlKind = "plan"
)

type AccountSubjectRouteRevision struct {
	ID               string           `json:"id"`
	TenantID         string           `json:"tenant_id"`
	ProjectID        string           `json:"project_id"`
	RouteID          string           `json:"route_id"`
	Revision         int              `json:"revision"`
	BillingAccountID string           `json:"billing_account_id"`
	UsageKind        UsageControlKind `json:"usage_kind"`
	UsageID          string           `json:"usage_id,omitempty"`
	ChangedAt        int64            `json:"changed_at"`
	ChangedBy        string           `json:"changed_by,omitempty"`
	ChangeReason     string           `json:"change_reason,omitempty"`
}

type UsageLimitPlanStatus string

const (
	UsageLimitPlanStatusDraft     UsageLimitPlanStatus = "draft"
	UsageLimitPlanStatusPublished UsageLimitPlanStatus = "published"
)

type UsageLimitPlan struct {
	ID              string               `json:"id"`
	TenantID        string               `json:"tenant_id"`
	ProjectID       string               `json:"project_id"`
	Name            string               `json:"name"`
	Title           string               `json:"title"`
	Description     string               `json:"description,omitempty"`
	Status          UsageLimitPlanStatus `json:"status"`
	CurrentRevision int                  `json:"current_revision"`
	CreatedAt       int64                `json:"created_at"`
	UpdatedAt       int64                `json:"updated_at"`
}

type UsageLimitPlanVersion struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	ProjectID   string               `json:"project_id"`
	PlanID      string               `json:"plan_id"`
	Name        string               `json:"name,omitempty"`
	Revision    int                  `json:"revision"`
	Status      UsageLimitPlanStatus `json:"status"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	PublishedAt int64                `json:"published_at,omitempty"`
	CreatedAt   int64                `json:"created_at"`
	Rules       []UsageLimitPlanRule `json:"rules,omitempty"`
}

type UsageLimitPlanRule struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	ProjectID     string          `json:"project_id"`
	PlanVersionID string          `json:"plan_version_id"`
	LimitKey      string          `json:"limit_key"`
	Currency      string          `json:"currency,omitempty"`
	Measure       string          `json:"measure"`
	SelectorType  string          `json:"selector_type"`
	SelectorValue string          `json:"selector_value,omitempty"`
	MetricScope   string          `json:"metric_scope,omitempty"`
	Period        string          `json:"period"`
	LimitAmount   decimal.Decimal `json:"limit_amount"`
	Timezone      string          `json:"timezone,omitempty"`
	Action        string          `json:"action,omitempty"`
	WebhookEvent  string          `json:"webhook_event,omitempty"`
}

type UsageLimitResetScope string

const (
	UsageLimitResetScopeProject        UsageLimitResetScope = "project"
	UsageLimitResetScopeBillingAccount UsageLimitResetScope = "billing_account"
	UsageLimitResetScopePlan           UsageLimitResetScope = "usage_limit_plan"
	UsageLimitResetScopeSubjectRoute   UsageLimitResetScope = "subject_route"
)

// UsageLimitResetOperation defines a scoped event-time boundary for the
// realtime usage-control projection. It never changes historical usage facts.
type UsageLimitResetOperation struct {
	ID                  string               `json:"id" db:"id"`
	TenantID            string               `json:"tenant_id" db:"tenant_id"`
	ProjectID           string               `json:"project_id" db:"project_id"`
	Scope               UsageLimitResetScope `json:"scope" db:"scope"`
	TargetID            string               `json:"target_id,omitempty" db:"target_id"`
	Generation          int64                `json:"generation" db:"generation"`
	EffectiveOccurredAt int64                `json:"effective_occurred_at" db:"effective_occurred_at"`
	RequestedBy         string               `json:"requested_by,omitempty" db:"requested_by"`
	Reason              string               `json:"reason,omitempty" db:"reason"`
	IdempotencyKey      string               `json:"-" db:"idempotency_key"`
	CreatedAt           int64                `json:"created_at" db:"created_at"`
}

type CreateUsageLimitResetRequest struct {
	Scope    UsageLimitResetScope `json:"scope,omitempty"`
	TargetID string               `json:"target_id,omitempty"`
	Reason   string               `json:"reason,omitempty"`
}

type CreateUsageLimitResetCommand struct {
	TenantID       string
	ProjectID      string
	Scope          UsageLimitResetScope
	TargetID       string
	RequestedBy    string
	Reason         string
	IdempotencyKey string
}

type CreateUsageLimitPlanRequest struct {
	Name        string                      `json:"name"`
	Title       string                      `json:"title"`
	Description string                      `json:"description,omitempty"`
	Rules       []UsageLimitPlanRuleRequest `json:"rules"`
}

type UsageLimitPlanRuleRequest struct {
	LimitKey      string          `json:"limit_key,omitempty"`
	Currency      string          `json:"currency,omitempty"`
	Measure       string          `json:"measure"`
	SelectorType  string          `json:"selector_type"`
	SelectorValue string          `json:"selector_value,omitempty"`
	MetricScope   string          `json:"metric_scope,omitempty"`
	Period        string          `json:"period"`
	LimitAmount   decimal.Decimal `json:"limit_amount"`
	Timezone      string          `json:"timezone,omitempty"`
	Action        string          `json:"action,omitempty"`
	WebhookEvent  string          `json:"webhook_event,omitempty"`
}

type CreateUsageLimitPlanCommand struct {
	TenantID    string
	ProjectID   string
	Name        string
	Title       string
	Description string
	Rules       []UsageLimitPlanRuleRequest
}

type UpdateUsageLimitPlanRequest struct {
	Title       string                      `json:"title"`
	Description string                      `json:"description,omitempty"`
	Rules       []UsageLimitPlanRuleRequest `json:"rules"`
}

type UpdateUsageLimitPlanCommand struct {
	TenantID    string
	ProjectID   string
	PlanID      string
	Title       string
	Description string
	Rules       []UsageLimitPlanRuleRequest
}

type UpdateAccountSubjectUsageControlRequest struct {
	UsageKind UsageControlKind `json:"usage_kind"`
	PlanID    string           `json:"plan_id,omitempty"`
}

type CacheInvalidationState string

const (
	CacheInvalidationStatePending    CacheInvalidationState = "pending"
	CacheInvalidationStateDelivering CacheInvalidationState = "delivering"
	CacheInvalidationStateDelivered  CacheInvalidationState = "delivered"
)

type ListAccountSubjectsRequest struct {
	AccountID string `json:"account_id"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type ListAccountSubjectsResponse struct {
	AccountSubjects []AccountSubject `json:"account_subjects"`
	TotalCount      int              `json:"total_count"`
	Limit           int              `json:"limit"`
	Offset          int              `json:"offset"`
}

type Wallet struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	ProjectID          string          `json:"project_id"`
	AccountID          string          `json:"account_id"`
	PrimarySubjectType string          `json:"primary_subject_type,omitempty"`
	PrimarySubjectID   string          `json:"primary_subject_id,omitempty"`
	Currency           string          `json:"currency"`
	Balance            decimal.Decimal `json:"balance"`
	AvailableBalance   decimal.Decimal `json:"available_balance"`
	CreditLimit        decimal.Decimal `json:"credit_limit"`
	ReconciledVersion  int64           `json:"reconciled_usage_version,omitempty"`
	Status             WalletStatus    `json:"status"`
	CreatedAt          int64           `json:"created_at"`
	UpdatedAt          int64           `json:"updated_at"`
}

type WalletLedgerEntry struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	ProjectID      string          `json:"project_id"`
	AccountID      string          `json:"account_id"`
	WalletID       string          `json:"wallet_id"`
	Currency       string          `json:"currency"`
	FundType       FundType        `json:"fund_type,omitempty"`
	FundID         string          `json:"fund_id,omitempty"`
	Direction      LedgerDirection `json:"direction"`
	Amount         decimal.Decimal `json:"amount"`
	BalanceAfter   decimal.Decimal `json:"balance_after"`
	SourceType     string          `json:"source_type,omitempty"`
	SourceID       string          `json:"source_id,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	UsageVersion   int64           `json:"usage_version,omitempty"`
	CreatedAt      int64           `json:"created_at"`
}

type VoucherGrant struct {
	ID                string             `json:"id"`
	TenantID          string             `json:"tenant_id"`
	ProjectID         string             `json:"project_id"`
	AccountID         string             `json:"account_id"`
	Currency          string             `json:"currency"`
	OriginalAmount    decimal.Decimal    `json:"original_amount"`
	RemainingAmount   decimal.Decimal    `json:"remaining_amount"`
	EffectiveFrom     int64              `json:"effective_from,omitempty"`
	ExpiresAt         int64              `json:"expires_at,omitempty"`
	Priority          int                `json:"priority,omitempty"`
	Status            VoucherGrantStatus `json:"status"`
	SourceType        string             `json:"source_type,omitempty"`
	SourceID          string             `json:"source_id,omitempty"`
	IdempotencyKey    string             `json:"idempotency_key,omitempty"`
	ReconciledVersion int64              `json:"reconciled_usage_version,omitempty"`
	CreatedAt         int64              `json:"created_at"`
	UpdatedAt         int64              `json:"updated_at"`
}

type RuleAccountStatus string

const (
	RuleAccountStatusActive   RuleAccountStatus = "active"
	RuleAccountStatusDisabled RuleAccountStatus = "disabled"
)

type RuleAccountLimitStatus string

const (
	RuleAccountLimitStatusActive   RuleAccountLimitStatus = "active"
	RuleAccountLimitStatusDisabled RuleAccountLimitStatus = "disabled"
)

// UsageLimitRuleKind identifies the immutable configuration owner of a usage
// limit. A subject route owns direct controls; a Plan version owns reusable
// controls. The owner ID is intentionally version-scoped for Plans so a
// published limit configuration remains reproducible.
type UsageLimitRuleKind string

const (
	UsageLimitRuleKindSubjectRoute UsageLimitRuleKind = "subject_route"
	UsageLimitRuleKindPlanVersion  UsageLimitRuleKind = "plan_version"
)

type RuleAccount struct {
	ID               string            `json:"id"`
	TenantID         string            `json:"tenant_id"`
	ProjectID        string            `json:"project_id"`
	BillingAccountID string            `json:"billing_account_id"`
	SubjectType      string            `json:"subject_type"`
	SubjectID        string            `json:"subject_id"`
	Status           RuleAccountStatus `json:"status"`
	CreatedAt        int64             `json:"created_at"`
	UpdatedAt        int64             `json:"updated_at"`
}

type RuleAccountLimit struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	ProjectID      string                 `json:"project_id"`
	RuleKind       UsageLimitRuleKind     `json:"rule_kind"`
	RuleID         string                 `json:"rule_id"`
	LimitKey       string                 `json:"limit_key,omitempty"`
	DimensionKey   string                 `json:"dimension_key,omitempty"`
	DimensionValue string                 `json:"dimension_value,omitempty"`
	Currency       string                 `json:"currency"`
	MetricScope    string                 `json:"metric_scope,omitempty"`
	SelectorType   string                 `json:"selector_type,omitempty"`
	SelectorValue  string                 `json:"selector_value,omitempty"`
	Measure        string                 `json:"measure,omitempty"`
	Period         string                 `json:"period"`
	LimitAmount    decimal.Decimal        `json:"limit_amount"`
	Timezone       string                 `json:"timezone,omitempty"`
	Action         string                 `json:"action,omitempty"`
	WebhookEvent   string                 `json:"webhook_event,omitempty"`
	Status         RuleAccountLimitStatus `json:"status"`
	CreatedAt      int64                  `json:"created_at"`
	UpdatedAt      int64                  `json:"updated_at"`
}

type ListUsageControlRulesRequest struct {
	AccountID string `json:"account_id"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type ListUsageControlRulesResponse struct {
	UsageControlRules []CreateUsageControlRuleResponse `json:"usage_control_rules"`
	TotalCount        int                              `json:"total_count"`
	Limit             int                              `json:"limit"`
	Offset            int                              `json:"offset"`
}

type CreateAccountRequest struct {
	Name               string `json:"name,omitempty"`
	PrimarySubjectType string `json:"primary_subject_type,omitempty"`
	PrimarySubjectID   string `json:"primary_subject_id,omitempty"`
}

type BindAccountSubjectRequest struct {
	AccountID     string `json:"account_id"`
	SubjectType   string `json:"subject_type"`
	SubjectID     string `json:"subject_id"`
	EffectiveFrom int64  `json:"effective_from,omitempty"`
}

type CreateWalletRequest struct {
	AccountID        string          `json:"account_id"`
	Currency         string          `json:"currency"`
	Balance          decimal.Decimal `json:"balance,omitempty"`
	AvailableBalance decimal.Decimal `json:"available_balance,omitempty"`
	CreditLimit      decimal.Decimal `json:"credit_limit,omitempty"`
}

type DebitWalletRequest struct {
	AccountID      string            `json:"account_id,omitempty"`
	SubjectType    string            `json:"subject_type,omitempty"`
	SubjectID      string            `json:"subject_id,omitempty"`
	Currency       string            `json:"currency"`
	Amount         decimal.Decimal   `json:"amount"`
	Control        map[string]string `json:"control,omitempty"`
	SourceType     string            `json:"source_type,omitempty"`
	SourceID       string            `json:"source_id,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	UsageVersion   int64             `json:"usage_version,omitempty"`
}

type DebitWalletResponse struct {
	Wallet        Wallet              `json:"wallet"`
	LedgerEntry   WalletLedgerEntry   `json:"ledger_entry"`
	LedgerEntries []WalletLedgerEntry `json:"ledger_entries,omitempty"`
}

type ListWalletLedgerEntriesRequest struct {
	AccountID string `json:"account_id,omitempty"`
	WalletID  string `json:"wallet_id,omitempty"`
	Currency  string `json:"currency,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type ListWalletLedgerEntriesResponse struct {
	LedgerEntries []WalletLedgerEntry `json:"ledger_entries"`
	TotalCount    int                 `json:"total_count"`
	Limit         int                 `json:"limit"`
	Offset        int                 `json:"offset"`
}

type CreditWalletRequest struct {
	AccountID      string          `json:"account_id,omitempty"`
	SubjectType    string          `json:"subject_type,omitempty"`
	SubjectID      string          `json:"subject_id,omitempty"`
	Currency       string          `json:"currency"`
	Amount         decimal.Decimal `json:"amount"`
	SourceType     string          `json:"source_type,omitempty"`
	SourceID       string          `json:"source_id,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}

type CreditWalletResponse struct {
	Wallet      Wallet            `json:"wallet"`
	LedgerEntry WalletLedgerEntry `json:"ledger_entry"`
}

type CreateVoucherGrantRequest struct {
	AccountID      string          `json:"account_id"`
	Currency       string          `json:"currency"`
	Amount         decimal.Decimal `json:"amount"`
	EffectiveFrom  int64           `json:"effective_from,omitempty"`
	ExpiresAt      int64           `json:"expires_at,omitempty"`
	Priority       int             `json:"priority,omitempty"`
	SourceType     string          `json:"source_type,omitempty"`
	SourceID       string          `json:"source_id,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}

type CreateVoucherGrantResponse struct {
	VoucherGrant VoucherGrant      `json:"voucher_grant"`
	LedgerEntry  WalletLedgerEntry `json:"ledger_entry"`
}

type ListVoucherGrantsResponse struct {
	VoucherGrants []VoucherGrant `json:"voucher_grants"`
	TotalCount    int            `json:"total_count"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

type RevokeVoucherGrantResponse struct {
	VoucherGrant  VoucherGrant        `json:"voucher_grant"`
	PendingAmount decimal.Decimal     `json:"pending_amount"`
	LedgerEntries []WalletLedgerEntry `json:"ledger_entries,omitempty"`
}

type SetWalletCreditLimitRequest struct {
	AccountID   string          `json:"account_id"`
	Currency    string          `json:"currency"`
	CreditLimit decimal.Decimal `json:"credit_limit"`
}

type ReadVirtualWalletRequest struct {
	AccountID   string `json:"account_id,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
	Currency    string `json:"currency"`
	ScopeKind   string `json:"scope_kind,omitempty"`
	ScopeKey    string `json:"scope_key,omitempty"`
	ScopeValue  string `json:"scope_value,omitempty"`
}

type VirtualWalletSnapshot struct {
	TenantID         string                       `json:"tenant_id"`
	ProjectID        string                       `json:"project_id"`
	AccountID        string                       `json:"account_id"`
	SubjectType      string                       `json:"subject_type,omitempty"`
	SubjectID        string                       `json:"subject_id,omitempty"`
	RuleAccountID    string                       `json:"rule_account_id,omitempty"`
	Currency         string                       `json:"currency"`
	Balance          decimal.Decimal              `json:"balance"`
	AvailableBalance decimal.Decimal              `json:"available_balance"`
	CashBalance      decimal.Decimal              `json:"cash_balance,omitempty"`
	CashAvailable    decimal.Decimal              `json:"cash_available_balance,omitempty"`
	VoucherAvailable decimal.Decimal              `json:"voucher_available_balance,omitempty"`
	CreditLimit      decimal.Decimal              `json:"credit_limit"`
	UsageVersion     int64                        `json:"usage_version,omitempty"`
	Scale            int                          `json:"scale,omitempty"`
	Limits           []VirtualWalletLimitSnapshot `json:"limits,omitempty"`
}

// VirtualWalletAmountSnapshot is the account-owned realtime balance view. It
// intentionally excludes subject usage-control state so balance reads do not
// need to resolve a subject route or usage-limit reset boundary.
type VirtualWalletAmountSnapshot struct {
	TenantID         string          `json:"tenant_id"`
	ProjectID        string          `json:"project_id"`
	AccountID        string          `json:"account_id"`
	Currency         string          `json:"currency"`
	Balance          decimal.Decimal `json:"balance"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
	CashBalance      decimal.Decimal `json:"cash_balance,omitempty"`
	CashAvailable    decimal.Decimal `json:"cash_available_balance,omitempty"`
	VoucherAvailable decimal.Decimal `json:"voucher_available_balance,omitempty"`
	CreditLimit      decimal.Decimal `json:"credit_limit"`
}

// VirtualWalletLimitsSnapshot is the subject-owned realtime usage-control
// view. The wallet amount is deliberately not repeated here.
type VirtualWalletLimitsSnapshot struct {
	TenantID      string                       `json:"tenant_id"`
	ProjectID     string                       `json:"project_id"`
	AccountID     string                       `json:"account_id"`
	SubjectType   string                       `json:"subject_type"`
	SubjectID     string                       `json:"subject_id"`
	RuleAccountID string                       `json:"rule_account_id,omitempty"`
	Currency      string                       `json:"currency"`
	ScopeKind     string                       `json:"scope_kind,omitempty"`
	ScopeKey      string                       `json:"scope_key,omitempty"`
	ScopeValue    string                       `json:"scope_value,omitempty"`
	Limits        []VirtualWalletLimitSnapshot `json:"limits"`
}

type VirtualWalletLimitSnapshot struct {
	RuleAccountID   string          `json:"rule_account_id,omitempty"`
	SubjectType     string          `json:"subject_type,omitempty"`
	SubjectID       string          `json:"subject_id,omitempty"`
	Currency        string          `json:"currency"`
	MetricScope     string          `json:"metric_scope"`
	SelectorType    string          `json:"selector_type,omitempty"`
	SelectorValue   string          `json:"selector_value,omitempty"`
	Measure         string          `json:"measure,omitempty"`
	Period          string          `json:"period"`
	WindowID        string          `json:"window_id"`
	UsedAmount      decimal.Decimal `json:"used_amount"`
	LimitAmount     decimal.Decimal `json:"limit_amount"`
	RemainingAmount decimal.Decimal `json:"remaining_amount"`
	Exceeded        bool            `json:"exceeded"`
	Action          string          `json:"action,omitempty"`
	WebhookEvent    string          `json:"webhook_event,omitempty"`
	UsageVersion    int64           `json:"-"`
}

type CreateRuleAccountLimitRequest struct {
	DimensionKey   string          `json:"dimension_key,omitempty"`
	DimensionValue string          `json:"dimension_value,omitempty"`
	Currency       string          `json:"currency"`
	MetricScope    string          `json:"metric_scope,omitempty"`
	SelectorType   string          `json:"selector_type,omitempty"`
	SelectorValue  string          `json:"selector_value,omitempty"`
	Measure        string          `json:"measure,omitempty"`
	Period         string          `json:"period"`
	LimitAmount    decimal.Decimal `json:"limit_amount"`
	Timezone       string          `json:"timezone,omitempty"`
	Action         string          `json:"action,omitempty"`
	WebhookEvent   string          `json:"webhook_event,omitempty"`
}

type CreateUsageControlRuleRequest struct {
	BillingAccountID string                          `json:"billing_account_id"`
	SubjectType      string                          `json:"subject_type"`
	SubjectID        string                          `json:"subject_id"`
	Currency         string                          `json:"currency,omitempty"`
	Limits           []CreateRuleAccountLimitRequest `json:"limits,omitempty"`
}

type CreateUsageControlRuleResponse struct {
	RuleAccount RuleAccount        `json:"rule_account"`
	Limits      []RuleAccountLimit `json:"limits"`
}

type DeleteRuleAccountLimitResponse struct {
	Deleted bool             `json:"deleted"`
	Limit   RuleAccountLimit `json:"limit"`
}
