package types

type CreateUsageReportCommand struct {
	ID          string
	TenantID    string
	ProjectID   string
	ReportID    string
	Name        string
	Description string
	SubjectType string
	BucketSize  string
	Filters     map[string]string
	GroupBy     []string
	Measures    []string
	Metrics     []string
	Status      RuleSetStatus
}
