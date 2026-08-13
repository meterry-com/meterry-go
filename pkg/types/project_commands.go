package types

type CreateProjectCommand struct {
	ID          string
	TenantID    string
	ProjectID   string
	Name        string
	Environment string
	Description string
	Status      RuleSetStatus
}
