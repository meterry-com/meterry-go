package types

import (
	"encoding/json"
	"strconv"

	"github.com/shopspring/decimal"
)

type WebhookEventType string

const (
	WebhookEventTypeWalletInsufficientBalance     WebhookEventType = "wallet.insufficient_balance"
	WebhookEventTypeWalletBalanceChanged          WebhookEventType = "wallet.balance_changed"
	WebhookEventTypeWalletBalanceThresholdCrossed WebhookEventType = "wallet.available_balance.threshold_crossed"
	WebhookEventTypeUsageLimitExhausted           WebhookEventType = "usage_limit.exhausted"
	WebhookEventTypeBudgetLeaseExhausted          WebhookEventType = "budget_lease.exhausted"
)

type WebhookBalanceThresholdOperator string

const (
	WebhookBalanceThresholdOperatorLT  WebhookBalanceThresholdOperator = "lt"
	WebhookBalanceThresholdOperatorLTE WebhookBalanceThresholdOperator = "lte"
	WebhookBalanceThresholdOperatorGT  WebhookBalanceThresholdOperator = "gt"
	WebhookBalanceThresholdOperatorGTE WebhookBalanceThresholdOperator = "gte"
)

type WebhookAvailableBalanceThreshold struct {
	ID        string                          `json:"id"`
	AccountID string                          `json:"account_id"`
	Currency  string                          `json:"currency"`
	Operator  WebhookBalanceThresholdOperator `json:"operator"`
	Threshold decimal.Decimal                 `json:"threshold"`
}

type WebhookEndpointEventConfigs struct {
	WalletAvailableBalanceThresholdCrossed []WebhookAvailableBalanceThreshold `json:"wallet.available_balance.threshold_crossed,omitempty"`
}

type WebhookEventStatus string

const (
	WebhookEventStatusPending    WebhookEventStatus = "pending"
	WebhookEventStatusDelivered  WebhookEventStatus = "delivered"
	WebhookEventStatusRetrying   WebhookEventStatus = "retrying"
	WebhookEventStatusFailed     WebhookEventStatus = "failed"
	WebhookEventStatusNoEndpoint WebhookEventStatus = "no_endpoint"
	WebhookEventStatusCanceled   WebhookEventStatus = "canceled"
)

type WebhookResourceType string

const (
	WebhookResourceTypeWallet      WebhookResourceType = "wallet"
	WebhookResourceTypeUsageLimit  WebhookResourceType = "usage_limit"
	WebhookResourceTypeBudgetLease WebhookResourceType = "budget_lease"
)

type WebhookTriggerStateValue string

const (
	WebhookTriggerStateNormal WebhookTriggerStateValue = "normal"
	WebhookTriggerStateOpen   WebhookTriggerStateValue = "open"
)

type WebhookEndpointStatus string

const (
	WebhookEndpointStatusActive   WebhookEndpointStatus = "active"
	WebhookEndpointStatusDisabled WebhookEndpointStatus = "disabled"
)

type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusSucceeded WebhookDeliveryStatus = "succeeded"
	WebhookDeliveryStatusFailed    WebhookDeliveryStatus = "failed"
	WebhookDeliveryStatusRetrying  WebhookDeliveryStatus = "retrying"
)

type WebhookEndpoint struct {
	ID           string                      `json:"id"`
	TenantID     string                      `json:"tenant_id"`
	ProjectID    string                      `json:"project_id"`
	Name         string                      `json:"name,omitempty"`
	URL          string                      `json:"url"`
	Secret       string                      `json:"-"`
	EventTypes   []string                    `json:"event_types"`
	EventConfigs WebhookEndpointEventConfigs `json:"event_configs,omitempty"`
	Status       WebhookEndpointStatus       `json:"status"`
	CreatedAt    int64                       `json:"created_at"`
	UpdatedAt    int64                       `json:"updated_at"`
}

type WebhookDelivery struct {
	ID                  string                `json:"id"`
	TenantID            string                `json:"tenant_id"`
	ProjectID           string                `json:"project_id"`
	EventID             string                `json:"event_id"`
	EndpointID          string                `json:"endpoint_id"`
	Attempt             int                   `json:"attempt"`
	Status              WebhookDeliveryStatus `json:"status"`
	RequestID           string                `json:"request_id,omitempty"`
	ResponseStatus      int                   `json:"response_status,omitempty"`
	ResponseBodyPreview string                `json:"response_body_preview,omitempty"`
	Error               string                `json:"error,omitempty"`
	DurationMS          int                   `json:"duration_ms,omitempty"`
	CreatedAt           int64                 `json:"created_at"`
}

type WebhookEvent struct {
	ID               string              `json:"id"`
	TenantID         string              `json:"tenant_id"`
	ProjectID        string              `json:"project_id"`
	Type             WebhookEventType    `json:"type"`
	EventName        string              `json:"event_name,omitempty"`
	TargetEndpointID string              `json:"target_endpoint_id,omitempty"`
	AccountID        string              `json:"account_id,omitempty"`
	SubjectType      string              `json:"subject_type,omitempty"`
	SubjectID        string              `json:"subject_id,omitempty"`
	Currency         string              `json:"currency,omitempty"`
	ResourceType     WebhookResourceType `json:"resource_type"`
	ResourceID       string              `json:"resource_id,omitempty"`
	TriggerKey       string              `json:"trigger_key"`
	DedupeKey        string              `json:"dedupe_key"`
	UsageEventID     string              `json:"usage_event_id,omitempty"`
	RawEventID       string              `json:"raw_event_id,omitempty"`
	UsageVersion     int64               `json:"usage_version,omitempty"`
	Payload          json.RawMessage     `json:"payload"`
	RequestBody      json.RawMessage     `json:"request_body,omitempty"`
	Status           WebhookEventStatus  `json:"status"`
	NextAttemptAt    int64               `json:"next_attempt_at,omitempty"`
	AttemptCount     int                 `json:"attempt_count,omitempty"`
	LastError        string              `json:"last_error,omitempty"`
	CreatedAt        int64               `json:"created_at"`
	UpdatedAt        int64               `json:"updated_at"`
}

type WebhookTriggerState struct {
	TenantID         string                   `json:"tenant_id"`
	ProjectID        string                   `json:"project_id"`
	TriggerKey       string                   `json:"trigger_key"`
	Type             WebhookEventType         `json:"type"`
	State            WebhookTriggerStateValue `json:"state"`
	LastEventID      string                   `json:"last_event_id,omitempty"`
	LastUsageVersion int64                    `json:"last_usage_version,omitempty"`
	Snapshot         json.RawMessage          `json:"snapshot,omitempty"`
	OpenedAt         int64                    `json:"opened_at,omitempty"`
	ClosedAt         int64                    `json:"closed_at,omitempty"`
	UpdatedAt        int64                    `json:"updated_at"`
}

type WebhookSignal struct {
	TenantID         string              `json:"tenant_id"`
	ProjectID        string              `json:"project_id"`
	Type             WebhookEventType    `json:"type"`
	EventName        string              `json:"event_name,omitempty"`
	TargetEndpointID string              `json:"target_endpoint_id,omitempty"`
	AccountID        string              `json:"account_id,omitempty"`
	SubjectType      string              `json:"subject_type,omitempty"`
	SubjectID        string              `json:"subject_id,omitempty"`
	Currency         string              `json:"currency,omitempty"`
	ResourceType     WebhookResourceType `json:"resource_type"`
	ResourceID       string              `json:"resource_id,omitempty"`
	TriggerKey       string              `json:"trigger_key"`
	DedupeKey        string              `json:"dedupe_key"`
	UsageEventID     string              `json:"usage_event_id,omitempty"`
	RawEventID       string              `json:"raw_event_id,omitempty"`
	UsageVersion     int64               `json:"usage_version,omitempty"`
	Payload          json.RawMessage     `json:"payload"`
	TriggeredAt      int64               `json:"triggered_at,omitempty"`
}

type ResolveWebhookTriggerCommand struct {
	TenantID     string
	ProjectID    string
	TriggerKey   string
	UsageVersion int64
	Snapshot     json.RawMessage
	ResolvedAt   int64
}

type ListWebhookEventsCommand struct {
	TenantID    string
	ProjectID   string
	Type        WebhookEventType
	EventName   string
	Status      WebhookEventStatus
	AccountID   string
	SubjectType string
	SubjectID   string
	AfterID     string
	Limit       int
}

type CreateWebhookEndpointRequest struct {
	Name         string                      `json:"name,omitempty"`
	URL          string                      `json:"url"`
	Secret       string                      `json:"secret,omitempty"`
	EventTypes   []string                    `json:"event_types"`
	EventConfigs WebhookEndpointEventConfigs `json:"event_configs,omitempty"`
	Status       WebhookEndpointStatus       `json:"status,omitempty"`
}

type UpdateWebhookEndpointRequest struct {
	Name         string                       `json:"name,omitempty"`
	URL          string                       `json:"url,omitempty"`
	Secret       string                       `json:"secret,omitempty"`
	EventTypes   []string                     `json:"event_types,omitempty"`
	EventConfigs *WebhookEndpointEventConfigs `json:"event_configs,omitempty"`
	Status       WebhookEndpointStatus        `json:"status,omitempty"`
}

type TestWebhookEndpointRequest struct {
	EndpointID string           `json:"endpoint_id,omitempty"`
	Name       string           `json:"name,omitempty"`
	URL        string           `json:"url,omitempty"`
	Secret     string           `json:"secret,omitempty"`
	EventType  WebhookEventType `json:"event_type,omitempty"`
	EventName  string           `json:"event_name,omitempty"`
}

type ListWebhookEndpointsResponse struct {
	Endpoints []WebhookEndpoint `json:"endpoints"`
}

type CreateWebhookEndpointResponse struct {
	Endpoint WebhookEndpoint `json:"endpoint"`
	Secret   string          `json:"secret,omitempty"`
}

type ListWebhookEventsResponse struct {
	Events []WebhookEvent `json:"events"`
}

type TestWebhookEndpointResponse struct {
	Delivery    WebhookDelivery `json:"delivery"`
	RequestBody json.RawMessage `json:"request_body,omitempty"`
}

type UpdateWebhookEventDeliveryStateCommand struct {
	TenantID      string
	ProjectID     string
	ID            string
	Status        WebhookEventStatus
	NextAttemptAt int64
	AttemptCount  int
	LastError     string
	UpdatedAt     int64
}

func WebhookSignaturePayload(timestamp int64, body []byte) string {
	return strconv.FormatInt(timestamp, 10) + "." + string(body)
}
