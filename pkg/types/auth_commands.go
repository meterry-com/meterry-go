package types

type CreateTenantAPIKeyCommand struct {
	TenantID        string
	ProjectID       string
	CreatedByUserID string
	KeyID           string
	Name            string
	Scopes          []string
}
