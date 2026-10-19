package sdk

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/meterry-com/meterry-go/pkg/types"
)

type ManagerClient struct {
	client *Client
}

// GrantSubject issues a short-lived subject JWT using a project-bound tenant
// API key with the subject_jwt:grant scope. The server derives tenant and
// project identity from the API key, rather than the configured project ID.
func (m *ManagerClient) GrantSubject(ctx context.Context, req types.GrantSubjectRequest) (*types.GrantSubjectResponse, error) {
	var out types.GrantSubjectResponse
	err := m.client.request(ctx, http.MethodPost, m.client.apiPath("subject", "grants"), nil, req, &out)
	return &out, err
}

type PublishResponse struct {
	Status string `json:"status"`
}

type ListWalletsRequest struct {
	AccountID   string
	SubjectType string
	SubjectID   string
}

type ListVoucherGrantsRequest struct {
	AccountID string
	Currency  string
	At        int64
	Limit     int
	Offset    int
	Expired   *bool
}

func (m *ManagerClient) CreateExtractorRuleSet(ctx context.Context, req types.CreateExtractorRuleSetRequest) (*types.ExtractorRuleSet, error) {
	var out types.ExtractorRuleSet
	err := m.client.request(ctx, http.MethodPost, m.projectPath("extractor-rule-sets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListExtractorRuleSets(ctx context.Context) (*types.ListExtractorRuleSetsResponse, error) {
	var out types.ListExtractorRuleSetsResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath("extractor-rule-sets"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) GetExtractorRuleSet(ctx context.Context, id string) (*types.ExtractorRuleSet, error) {
	var out types.ExtractorRuleSet
	err := m.client.request(ctx, http.MethodGet, m.projectPath("extractor-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) PublishExtractorRuleSet(ctx context.Context, id string) (*PublishResponse, error) {
	var out PublishResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("extractor-rule-sets", id, "publish"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) DeleteExtractorRuleSet(ctx context.Context, id string) (*types.DeleteRuleSetResponse, error) {
	var out types.DeleteRuleSetResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath("extractor-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreatePricingRuleSet(ctx context.Context, req types.CreatePricingRuleSetRequest) (*types.PricingRuleSet, error) {
	var out types.PricingRuleSet
	err := m.client.request(ctx, http.MethodPost, m.projectPath("pricing-rule-sets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListPricingRuleSets(ctx context.Context) (*types.ListPricingRuleSetsResponse, error) {
	var out types.ListPricingRuleSetsResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath("pricing-rule-sets"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) PublishPricingRuleSet(ctx context.Context, id string) (*PublishResponse, error) {
	var out PublishResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("pricing-rule-sets", id, "publish"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) DeletePricingRuleSet(ctx context.Context, id string) (*types.DeleteRuleSetResponse, error) {
	var out types.DeleteRuleSetResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath("pricing-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreateAccount(ctx context.Context, req types.CreateAccountRequest) (*types.Account, error) {
	var out types.Account
	err := m.client.request(ctx, http.MethodPost, m.projectPath("accounts"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListAccounts(ctx context.Context) ([]types.Account, error) {
	var out struct {
		Accounts []types.Account `json:"accounts"`
	}
	err := m.client.request(ctx, http.MethodGet, m.projectPath("accounts"), nil, nil, &out)
	return out.Accounts, err
}

func (m *ManagerClient) BindSubject(ctx context.Context, req types.BindAccountSubjectRequest) (*types.AccountSubject, error) {
	var out types.AccountSubject
	err := m.client.request(ctx, http.MethodPost, m.projectPath("account-subjects"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListAccountSubjects(ctx context.Context, accountID string) ([]types.AccountSubject, error) {
	const pageSize = 500
	result := make([]types.AccountSubject, 0)
	for offset := 0; ; offset += pageSize {
		query := url.Values{}
		if accountID != "" {
			query.Set("account_id", accountID)
		}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		var out types.ListAccountSubjectsResponse
		if err := m.client.request(ctx, http.MethodGet, m.projectPath("account-subjects"), query, nil, &out); err != nil {
			return nil, err
		}
		result = append(result, out.AccountSubjects...)
		if len(out.AccountSubjects) < pageSize || offset+len(out.AccountSubjects) >= out.TotalCount {
			return result, nil
		}
	}
}

func (m *ManagerClient) CreateWallet(ctx context.Context, req types.CreateWalletRequest) (*types.Wallet, error) {
	var out types.Wallet
	err := m.client.request(ctx, http.MethodPost, m.projectPath("wallets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListWallets(ctx context.Context, req ListWalletsRequest) ([]types.Wallet, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	var out struct {
		Wallets []types.Wallet `json:"wallets"`
	}
	err := m.client.request(ctx, http.MethodGet, m.projectPath("wallets"), query, nil, &out)
	return out.Wallets, err
}

func (m *ManagerClient) ReadVirtualWalletAmount(ctx context.Context, req types.ReadVirtualWalletRequest) (*types.VirtualWalletAmountSnapshot, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	addQuery(query, "currency", req.Currency)
	var out types.VirtualWalletAmountSnapshot
	err := m.client.request(ctx, http.MethodGet, m.projectPath("wallets", "realtime", "amount"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) ReadSubjectRealtimeAmount(ctx context.Context, subjectType, subjectID, currency string) (*types.VirtualWalletAmountSnapshot, error) {
	return m.ReadVirtualWalletAmount(ctx, types.ReadVirtualWalletRequest{
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Currency:    currency,
	})
}

func (m *ManagerClient) ReadSubjectRealtimeLimits(ctx context.Context, req types.ReadSubjectRealtimeLimitsRequest) (*types.SubjectRealtimeLimitsSnapshot, error) {
	query := url.Values{}
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	var out types.SubjectRealtimeLimitsSnapshot
	err := m.client.request(ctx, http.MethodGet, m.projectPath("subject", "realtime", "limits"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) DebitWallet(ctx context.Context, req types.DebitWalletRequest) (*types.DebitWalletResponse, error) {
	var out types.DebitWalletResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("wallets", "debit"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) CreditWallet(ctx context.Context, req types.CreditWalletRequest) (*types.CreditWalletResponse, error) {
	var out types.CreditWalletResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("wallets", "credit"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) TopUpWallet(ctx context.Context, req types.CreditWalletRequest) (*types.CreditWalletResponse, error) {
	return m.CreditWallet(ctx, req)
}

func (m *ManagerClient) SetWalletCreditLimit(ctx context.Context, req types.SetWalletCreditLimitRequest) (*types.Wallet, error) {
	var out types.Wallet
	err := m.client.request(ctx, http.MethodPost, m.projectPath("wallets", "credit-limit"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListWalletLedgerEntries(ctx context.Context, req types.ListWalletLedgerEntriesRequest) (*types.ListWalletLedgerEntriesResponse, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "wallet_id", req.WalletID)
	addQuery(query, "currency", req.Currency)
	addIntQuery(query, "limit", req.Limit)
	addIntQuery(query, "offset", req.Offset)
	var out types.ListWalletLedgerEntriesResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath("wallet-ledger-entries"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreateVoucherGrant(ctx context.Context, req types.CreateVoucherGrantRequest) (*types.CreateVoucherGrantResponse, error) {
	var out types.CreateVoucherGrantResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("voucher-grants"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListVoucherGrants(ctx context.Context, req ListVoucherGrantsRequest) (*types.ListVoucherGrantsResponse, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "currency", req.Currency)
	addInt64Query(query, "at", req.At)
	addIntQuery(query, "limit", req.Limit)
	addIntQuery(query, "offset", req.Offset)
	if req.Expired != nil {
		if *req.Expired {
			query.Set("expired", "true")
		} else {
			query.Set("expired", "false")
		}
	}
	var out types.ListVoucherGrantsResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath("voucher-grants"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) RevokeVoucherGrant(ctx context.Context, id string) (*types.RevokeVoucherGrantResponse, error) {
	var out types.RevokeVoucherGrantResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("voucher-grants", id, "revoke"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) SetSubjectRouteAndUsageLimits(ctx context.Context, req types.CreateUsageControlRuleRequest) (*types.CreateUsageControlRuleResponse, error) {
	return m.CreateUsageControlRule(ctx, req)
}

func (m *ManagerClient) CreateUsageControlRule(ctx context.Context, req types.CreateUsageControlRuleRequest) (*types.CreateUsageControlRuleResponse, error) {
	var out types.CreateUsageControlRuleResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath("usage-control-rules"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListUsageControlRules(ctx context.Context, accountID string) ([]types.CreateUsageControlRuleResponse, error) {
	return m.listUsageControlRules(ctx, types.ListUsageControlRulesRequest{AccountID: accountID})
}

func (m *ManagerClient) listUsageControlRules(ctx context.Context, req types.ListUsageControlRulesRequest) ([]types.CreateUsageControlRuleResponse, error) {
	const pageSize = 500
	result := make([]types.CreateUsageControlRuleResponse, 0)
	for offset := 0; ; offset += pageSize {
		query := url.Values{}
		addQuery(query, "account_id", req.AccountID)
		addQuery(query, "subject_type", req.SubjectType)
		addQuery(query, "subject_id", req.SubjectID)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		var out types.ListUsageControlRulesResponse
		if err := m.client.request(ctx, http.MethodGet, m.projectPath("usage-control-rules"), query, nil, &out); err != nil {
			return nil, err
		}
		result = append(result, out.UsageControlRules...)
		if len(out.UsageControlRules) < pageSize || offset+len(out.UsageControlRules) >= out.TotalCount {
			return result, nil
		}
	}
}

func (m *ManagerClient) DeleteRuleAccountLimit(ctx context.Context, id string) (*types.DeleteRuleAccountLimitResponse, error) {
	var out types.DeleteRuleAccountLimitResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath("usage-control-rules", "limits", id), nil, nil, &out)
	return &out, err
}

// CreateBudgetLease creates a one-time agent-run budget bound to an
// existing stable subject route. Use the returned ID in x-usage-control on
// every usage event for that run.
func (m *ManagerClient) CreateBudgetLease(ctx context.Context, req types.CreateBudgetLeaseRequest) (*types.BudgetLease, error) {
	var out types.BudgetLease
	err := m.client.request(ctx, http.MethodPost, m.projectPath("budget-leases"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) projectPath(parts ...string) string {
	return m.client.projectPath(parts...)
}

// ImportSubjectUsage imports or replaces the absolute used baseline of one
// custom limit. Reuse the request and idempotency key when retrying.
func (m *ManagerClient) ImportSubjectUsage(ctx context.Context, req types.ImportSubjectUsageRequest) (*types.SubjectUsageImport, error) {
	var out types.SubjectUsageImport
	err := m.client.request(ctx, http.MethodPost, m.projectPath("subject", "usage-imports"), nil, req, &out)
	return &out, err
}

// GetSubjectUsageImport returns the accepted import and its application status.
func (m *ManagerClient) GetSubjectUsageImport(ctx context.Context, id string) (*types.SubjectUsageImport, error) {
	var out types.SubjectUsageImport
	err := m.client.request(ctx, http.MethodGet, m.projectPath("subject", "usage-imports", id), nil, nil, &out)
	return &out, err
}

// UpdateUsageControlLimit updates one custom limit. Cap changes preserve used;
// period changes reset only the edited limit.
func (m *ManagerClient) UpdateUsageControlLimit(ctx context.Context, id string, req types.UpdateUsageControlLimitRequest) (*types.RuleAccountLimit, error) {
	var out types.RuleAccountLimit
	err := m.client.request(ctx, http.MethodPatch, m.projectPath("usage-control-rules", "limits", id), nil, req, &out)
	return &out, err
}

// ResetUsageControlLimit resets only the given limit, overriding Scope and
// TargetID in req. Manual resets do not emit recovered webhooks.
func (m *ManagerClient) ResetUsageControlLimit(ctx context.Context, id string, req types.CreateUsageLimitResetRequest) (*types.UsageLimitResetOperation, error) {
	var out types.UsageLimitResetOperation
	req.Scope = types.UsageLimitResetScopeLimit
	req.TargetID = id
	err := m.client.request(ctx, http.MethodPost, m.projectPath("usage-limit-resets"), nil, req, &out)
	return &out, err
}
