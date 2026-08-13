package types

import "encoding/json"

type IngestEventCommand struct {
	RawEventID         string
	TenantID           string
	ProjectID          string
	Source             string
	ExternalEventID    string
	IdempotencyKey     string
	ExtractorRuleSetID string
	SubjectType        string
	SubjectID          string
	OccurredAt         int64
	IngestedAt         int64
	RawJSON            json.RawMessage
	XBilling           *XBilling
	XUsageControl      *XUsageControl
	Metadata           map[string]string
}

type IngestEventMessage struct {
	RawEventID         string            `json:"raw_event_id,omitempty"`
	TenantID           string            `json:"tenant_id"`
	ProjectID          string            `json:"project_id,omitempty"`
	Source             string            `json:"source"`
	ExternalEventID    string            `json:"external_event_id,omitempty"`
	IdempotencyKey     string            `json:"idempotency_key,omitempty"`
	ExtractorRuleSetID string            `json:"extractor_rule_set_id,omitempty"`
	SubjectType        string            `json:"subject_type,omitempty"`
	SubjectID          string            `json:"subject_id,omitempty"`
	OccurredAt         int64             `json:"occurred_at"`
	IngestedAt         int64             `json:"ingested_at,omitempty"`
	RawJSON            json.RawMessage   `json:"raw_json"`
	XBilling           *XBilling         `json:"x-billing,omitempty"`
	XUsageControl      *XUsageControl    `json:"x-usage-control,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}
