package types

type TenantAPIKeyStatus string

const (
	TenantAPIKeyStatusActive   TenantAPIKeyStatus = "active"
	TenantAPIKeyStatusDisabled TenantAPIKeyStatus = "disabled"
)

const (
	TenantAPIKeyScopeUsageIngest       = "usage:ingest"
	TenantAPIKeyScopeUsageRead         = "usage:read"
	TenantAPIKeyScopeAnalyticsRead     = "analytics:read"
	TenantAPIKeyScopeAnalyticsWrite    = "analytics:write"
	TenantAPIKeyScopeRulesRead         = "rules:read"
	TenantAPIKeyScopeRulesWrite        = "rules:write"
	TenantAPIKeyScopeRulesTest         = "rules:test"
	TenantAPIKeyScopeAccountsRead      = "accounts:read"
	TenantAPIKeyScopeAccountsWrite     = "accounts:write"
	TenantAPIKeyScopeWalletsRead       = "wallets:read"
	TenantAPIKeyScopeWalletsWrite      = "wallets:write"
	TenantAPIKeyScopeUsageControlWrite = "usage_control:write"
	TenantAPIKeyScopeTenantAccessWrite = "tenant_access:write"
	TenantAPIKeyScopeTenantAdmin       = "tenant:admin"
)

type TenantAPIKey struct {
	ID              string             `json:"id"`
	KeyID           string             `json:"key_id"`
	TenantID        string             `json:"tenant_id"`
	ProjectID       string             `json:"project_id,omitempty"`
	Name            string             `json:"name,omitempty"`
	SecretHash      string             `json:"-"`
	Scopes          []string           `json:"scopes,omitempty"`
	Status          TenantAPIKeyStatus `json:"status"`
	CreatedByUserID string             `json:"created_by_user_id,omitempty"`
	LastUsedAt      int64              `json:"last_used_at,omitempty"`
	CreatedAt       int64              `json:"created_at"`
	UpdatedAt       int64              `json:"updated_at"`
}

type CreateTenantAPIKeyRequest struct {
	TenantID        string   `json:"tenant_id,omitempty"`
	KeyID           string   `json:"key_id,omitempty"`
	Name            string   `json:"name,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`
	CreatedByUserID string   `json:"created_by_user_id,omitempty"`
}

type CreateTenantAPIKeyResponse struct {
	APIKey TenantAPIKey `json:"api_key"`
	Secret string       `json:"secret"`
}

type ListTenantAPIKeysResponse struct {
	APIKeys []TenantAPIKey `json:"api_keys"`
}

type DeleteTenantAPIKeyResponse struct {
	Deleted bool `json:"deleted"`
}
