package types

import "github.com/shopspring/decimal"

type CreateAccountCommand struct {
	TenantID           string
	ProjectID          string
	Name               string
	PrimarySubjectType string
	PrimarySubjectID   string
}

type BindAccountSubjectCommand struct {
	TenantID      string
	ProjectID     string
	AccountID     string
	SubjectType   string
	SubjectID     string
	EffectiveFrom int64
}

type CreateWalletCommand struct {
	TenantID         string
	ProjectID        string
	AccountID        string
	Currency         string
	Balance          decimal.Decimal
	AvailableBalance decimal.Decimal
	CreditLimit      decimal.Decimal
}

type DebitWalletCommand struct {
	TenantID       string
	ProjectID      string
	AccountID      string
	SubjectType    string
	SubjectID      string
	Currency       string
	Amount         decimal.Decimal
	Control        map[string]string
	SourceType     string
	SourceID       string
	IdempotencyKey string
	UsageVersion   int64
}

type CreditWalletCommand struct {
	TenantID       string
	ProjectID      string
	AccountID      string
	SubjectType    string
	SubjectID      string
	Currency       string
	Amount         decimal.Decimal
	SourceType     string
	SourceID       string
	Remark         string
	IdempotencyKey string
}

type ListWalletLedgerEntriesCommand struct {
	TenantID  string
	ProjectID string
	AccountID string
	WalletID  string
	Currency  string
	Limit     int
	Offset    int
}

type CreateVoucherGrantCommand struct {
	TenantID       string
	ProjectID      string
	AccountID      string
	Currency       string
	Amount         decimal.Decimal
	EffectiveFrom  int64
	ExpiresAt      int64
	Priority       int
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Remark         string
}

type ListVoucherGrantsCommand struct {
	TenantID  string
	ProjectID string
	AccountID string
	Currency  string
	At        int64
	Limit     int
	Offset    int
	Expired   *bool
}

type RevokeVoucherGrantCommand struct {
	TenantID  string
	ProjectID string
	ID        string
}

type SetWalletCreditLimitCommand struct {
	TenantID    string
	ProjectID   string
	AccountID   string
	Currency    string
	CreditLimit decimal.Decimal
}

type ReadVirtualWalletCommand struct {
	TenantID    string
	ProjectID   string
	AccountID   string
	SubjectType string
	SubjectID   string
	Currency    string
	ScopeKind   string
	ScopeKey    string
	ScopeValue  string
}

type CreateUsageControlRuleCommand struct {
	TenantID         string
	ProjectID        string
	BillingAccountID string
	SubjectType      string
	SubjectID        string
	Currency         string
	Limits           []CreateRuleAccountLimitRequest
}
