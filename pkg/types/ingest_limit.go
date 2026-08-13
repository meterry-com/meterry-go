package types

const (
	CommercialStatusFree       = "free"
	CommercialStatusCommercial = "commercial"
	CommercialStatusEnterprise = "enterprise"
)

type IngestLimitProfile struct {
	CommercialStatus         string  `json:"commercial_status"`
	Bypass                   bool    `json:"bypass"`
	TenantRequestsPerSecond  float64 `json:"tenant_requests_per_second"`
	TenantRequestBurst       int     `json:"tenant_request_burst"`
	ProjectRequestsPerSecond float64 `json:"project_requests_per_second"`
	ProjectRequestBurst      int     `json:"project_request_burst"`
}
