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

func (m *ManagerClient) CreateExtractorRuleSet(ctx context.Context, projectID string, req types.CreateExtractorRuleSetRequest) (*types.ExtractorRuleSet, error) {
	var out types.ExtractorRuleSet
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "extractor-rule-sets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListExtractorRuleSets(ctx context.Context, projectID string) (*types.ListExtractorRuleSetsResponse, error) {
	var out types.ListExtractorRuleSetsResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "extractor-rule-sets"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) GetExtractorRuleSet(ctx context.Context, projectID, id string) (*types.ExtractorRuleSet, error) {
	var out types.ExtractorRuleSet
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "extractor-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) PublishExtractorRuleSet(ctx context.Context, projectID, id string) (*PublishResponse, error) {
	var out PublishResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "extractor-rule-sets", id, "publish"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) DeleteExtractorRuleSet(ctx context.Context, projectID, id string) (*types.DeleteRuleSetResponse, error) {
	var out types.DeleteRuleSetResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath(projectID, "extractor-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreatePricingRuleSet(ctx context.Context, projectID string, req types.CreatePricingRuleSetRequest) (*types.PricingRuleSet, error) {
	var out types.PricingRuleSet
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "pricing-rule-sets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListPricingRuleSets(ctx context.Context, projectID string) (*types.ListPricingRuleSetsResponse, error) {
	var out types.ListPricingRuleSetsResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "pricing-rule-sets"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) PublishPricingRuleSet(ctx context.Context, projectID, id string) (*PublishResponse, error) {
	var out PublishResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "pricing-rule-sets", id, "publish"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) DeletePricingRuleSet(ctx context.Context, projectID, id string) (*types.DeleteRuleSetResponse, error) {
	var out types.DeleteRuleSetResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath(projectID, "pricing-rule-sets", id), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreateAccountForProject(ctx context.Context, projectID string, req types.CreateAccountRequest) (*types.Account, error) {
	var out types.Account
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "accounts"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListAccountsForProject(ctx context.Context, projectID string) ([]types.Account, error) {
	var out struct {
		Accounts []types.Account `json:"accounts"`
	}
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "accounts"), nil, nil, &out)
	return out.Accounts, err
}

func (m *ManagerClient) BindSubjectForProject(ctx context.Context, projectID string, req types.BindAccountSubjectRequest) (*types.AccountSubject, error) {
	var out types.AccountSubject
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "account-subjects"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListAccountSubjectsForProject(ctx context.Context, projectID, accountID string) ([]types.AccountSubject, error) {
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
		if err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "account-subjects"), query, nil, &out); err != nil {
			return nil, err
		}
		result = append(result, out.AccountSubjects...)
		if len(out.AccountSubjects) < pageSize || offset+len(out.AccountSubjects) >= out.TotalCount {
			return result, nil
		}
	}
}

func (m *ManagerClient) CreateWalletForProject(ctx context.Context, projectID string, req types.CreateWalletRequest) (*types.Wallet, error) {
	var out types.Wallet
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "wallets"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListWalletsForProject(ctx context.Context, projectID string, req ListWalletsRequest) ([]types.Wallet, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	var out struct {
		Wallets []types.Wallet `json:"wallets"`
	}
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "wallets"), query, nil, &out)
	return out.Wallets, err
}

func (m *ManagerClient) ReadVirtualWalletAmountForProject(ctx context.Context, projectID string, req types.ReadVirtualWalletRequest) (*types.VirtualWalletAmountSnapshot, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	addQuery(query, "currency", req.Currency)
	var out types.VirtualWalletAmountSnapshot
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "wallets", "realtime", "amount"), query, nil, &out)
	return &out, err
}

// ReadRealtimeWalletLimitsForProject preserves the legacy wallet-limits
// snapshot shape while sourcing data from the subject realtime limits API.
// Deprecated: use ReadSubjectRealtimeLimitsForProject.
func (m *ManagerClient) ReadRealtimeWalletLimitsForProject(ctx context.Context, projectID string, req types.ReadVirtualWalletRequest) (*types.VirtualWalletLimitsSnapshot, error) {
	snapshot, err := m.ReadSubjectRealtimeLimitsForProject(ctx, projectID, types.ReadSubjectRealtimeLimitsRequest{
		SubjectType: req.SubjectType,
		SubjectID:   req.SubjectID,
	})
	if err != nil {
		return nil, err
	}
	return subjectRealtimeLimitsToVirtualWalletLimitsSnapshot(snapshot, req), nil
}

// ReadVirtualWalletLimitsForProject preserves the legacy wallet-limits
// snapshot shape while sourcing data from the subject realtime limits API.
// Deprecated: use ReadSubjectRealtimeLimitsForProject.
func (m *ManagerClient) ReadVirtualWalletLimitsForProject(ctx context.Context, projectID string, req types.ReadVirtualWalletRequest) (*types.VirtualWalletLimitsSnapshot, error) {
	return m.ReadRealtimeWalletLimitsForProject(ctx, projectID, req)
}

func (m *ManagerClient) ReadSubjectRealtimeLimitsForProject(ctx context.Context, projectID string, req types.ReadSubjectRealtimeLimitsRequest) (*types.SubjectRealtimeLimitsSnapshot, error) {
	query := url.Values{}
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	var out types.SubjectRealtimeLimitsSnapshot
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "subject", "realtime", "limits"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) DebitWalletForProject(ctx context.Context, projectID string, req types.DebitWalletRequest) (*types.DebitWalletResponse, error) {
	var out types.DebitWalletResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "wallets", "debit"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) CreditWalletForProject(ctx context.Context, projectID string, req types.CreditWalletRequest) (*types.CreditWalletResponse, error) {
	var out types.CreditWalletResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "wallets", "credit"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) TopUpWalletForProject(ctx context.Context, projectID string, req types.CreditWalletRequest) (*types.CreditWalletResponse, error) {
	return m.CreditWalletForProject(ctx, projectID, req)
}

func (m *ManagerClient) SetWalletCreditLimitForProject(ctx context.Context, projectID string, req types.SetWalletCreditLimitRequest) (*types.Wallet, error) {
	var out types.Wallet
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "wallets", "credit-limit"), nil, req, &out)
	return &out, err
}

func subjectRealtimeLimitsToVirtualWalletLimitsSnapshot(snapshot *types.SubjectRealtimeLimitsSnapshot, req types.ReadVirtualWalletRequest) *types.VirtualWalletLimitsSnapshot {
	if snapshot == nil {
		return nil
	}
	out := &types.VirtualWalletLimitsSnapshot{
		TenantID:      snapshot.TenantID,
		ProjectID:     snapshot.ProjectID,
		AccountID:     snapshot.AccountID,
		SubjectType:   snapshot.SubjectType,
		SubjectID:     snapshot.SubjectID,
		RuleAccountID: snapshot.RuleAccountID,
		Currency:      req.Currency,
		ScopeKind:     req.ScopeKind,
		ScopeKey:      req.ScopeKey,
		ScopeValue:    req.ScopeValue,
	}
	if out.Currency == "" {
		for _, item := range snapshot.Limits {
			if item.Currency != "" {
				out.Currency = item.Currency
				break
			}
		}
	}
	out.Limits = make([]types.VirtualWalletLimitSnapshot, 0, len(snapshot.Limits))
	for _, item := range snapshot.Limits {
		if req.Currency != "" && item.Currency != "" && item.Currency != req.Currency {
			continue
		}
		out.Limits = append(out.Limits, types.VirtualWalletLimitSnapshot{
			RuleAccountID:   item.RuleAccountID,
			SubjectType:     item.SubjectType,
			SubjectID:       item.SubjectID,
			Currency:        item.Currency,
			MetricScope:     item.MetricScope,
			SelectorType:    item.SelectorType,
			SelectorValue:   item.SelectorValue,
			Measure:         item.Measure,
			Period:          item.Period,
			WindowID:        item.WindowID,
			UsedAmount:      item.UsedAmount,
			LimitAmount:     item.LimitAmount,
			RemainingAmount: item.RemainingAmount,
			Exceeded:        item.Exceeded,
			Action:          item.Action,
			WebhookEvent:    item.WebhookEvent,
		})
	}
	return out
}

func (m *ManagerClient) ListWalletLedgerEntriesForProject(ctx context.Context, projectID string, req types.ListWalletLedgerEntriesRequest) (*types.ListWalletLedgerEntriesResponse, error) {
	query := url.Values{}
	addQuery(query, "account_id", req.AccountID)
	addQuery(query, "wallet_id", req.WalletID)
	addQuery(query, "currency", req.Currency)
	addIntQuery(query, "limit", req.Limit)
	addIntQuery(query, "offset", req.Offset)
	var out types.ListWalletLedgerEntriesResponse
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "wallet-ledger-entries"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) CreateVoucherGrantForProject(ctx context.Context, projectID string, req types.CreateVoucherGrantRequest) (*types.CreateVoucherGrantResponse, error) {
	var out types.CreateVoucherGrantResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "voucher-grants"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListVoucherGrantsForProject(ctx context.Context, projectID string, req ListVoucherGrantsRequest) (*types.ListVoucherGrantsResponse, error) {
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
	err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "voucher-grants"), query, nil, &out)
	return &out, err
}

func (m *ManagerClient) RevokeVoucherGrantForProject(ctx context.Context, projectID, id string) (*types.RevokeVoucherGrantResponse, error) {
	var out types.RevokeVoucherGrantResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "voucher-grants", id, "revoke"), nil, nil, &out)
	return &out, err
}

func (m *ManagerClient) SetSubjectRouteAndUsageLimitsForProject(ctx context.Context, projectID string, req types.CreateUsageControlRuleRequest) (*types.CreateUsageControlRuleResponse, error) {
	return m.CreateUsageControlRuleForProject(ctx, projectID, req)
}

func (m *ManagerClient) CreateUsageControlRuleForProject(ctx context.Context, projectID string, req types.CreateUsageControlRuleRequest) (*types.CreateUsageControlRuleResponse, error) {
	var out types.CreateUsageControlRuleResponse
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "usage-control-rules"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) ListUsageControlRulesForProject(ctx context.Context, projectID, accountID string) ([]types.CreateUsageControlRuleResponse, error) {
	const pageSize = 500
	result := make([]types.CreateUsageControlRuleResponse, 0)
	for offset := 0; ; offset += pageSize {
		query := url.Values{}
		addQuery(query, "account_id", accountID)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		var out types.ListUsageControlRulesResponse
		if err := m.client.request(ctx, http.MethodGet, m.projectPath(projectID, "usage-control-rules"), query, nil, &out); err != nil {
			return nil, err
		}
		result = append(result, out.UsageControlRules...)
		if len(out.UsageControlRules) < pageSize || offset+len(out.UsageControlRules) >= out.TotalCount {
			return result, nil
		}
	}
}

func (m *ManagerClient) DeleteRuleAccountLimitForProject(ctx context.Context, projectID, id string) (*types.DeleteRuleAccountLimitResponse, error) {
	var out types.DeleteRuleAccountLimitResponse
	err := m.client.request(ctx, http.MethodDelete, m.projectPath(projectID, "usage-control-rules", "limits", id), nil, nil, &out)
	return &out, err
}

// CreateBudgetLeaseForProject creates a one-time agent-run budget bound to an
// existing stable subject route. Use the returned ID in x-usage-control on
// every usage event for that run.
func (m *ManagerClient) CreateBudgetLeaseForProject(ctx context.Context, projectID string, req types.CreateBudgetLeaseRequest) (*types.BudgetLease, error) {
	var out types.BudgetLease
	err := m.client.request(ctx, http.MethodPost, m.projectPath(projectID, "budget-leases"), nil, req, &out)
	return &out, err
}

func (m *ManagerClient) projectPath(projectID string, parts ...string) string {
	values := append([]string{"projects", projectID}, parts...)
	return m.client.apiPath(values...)
}
