package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/meterry-com/meterry-go/pkg/types"
)

type IngestClient struct {
	client *Client
}

type IngestEvent struct {
	Source           string
	ExternalEventID  string
	IdempotencyKey   string
	SubjectType      string
	SubjectID        string
	OccurredAt       int64
	RawJSON          any
	XBilling         *types.XBilling
	PricingRuleSetID string
	Metadata         map[string]string
}

type BatchIngestEvent struct {
	ExtractorRuleSetID string
	IngestEvent
}

func (i *IngestClient) IngestRawEvent(ctx context.Context, extractorRuleSetID string, req types.IngestEventRequest) (*types.IngestEventResponse, error) {
	if extractorRuleSetID == "" {
		return nil, errors.New("extractor rule set id is required")
	}
	endpoint := i.client.projectPath("extractor-rule-sets", extractorRuleSetID, "events", "ingest")
	var out types.IngestEventResponse
	err := i.client.request(ctx, http.MethodPost, endpoint, nil, req, &out)
	return &out, err
}

func (i *IngestClient) Ingest(ctx context.Context, extractorRuleSetID string, event IngestEvent) (*types.IngestEventResponse, error) {
	raw, err := BuildRawJSON(event.RawJSON, nil, event.PricingRuleSetID)
	if err != nil {
		return nil, err
	}
	return i.IngestRawEvent(ctx, extractorRuleSetID, types.IngestEventRequest{
		Source:          event.Source,
		ExternalEventID: event.ExternalEventID,
		IdempotencyKey:  event.IdempotencyKey,
		SubjectType:     event.SubjectType,
		SubjectID:       event.SubjectID,
		OccurredAt:      event.OccurredAt,
		RawJSON:         raw,
		XBilling:        event.XBilling,
		Metadata:        event.Metadata,
	})
}

func (i *IngestClient) BatchIngestRawEvents(ctx context.Context, req types.BatchIngestEventsRequest) (*types.BatchIngestEventsResponse, error) {
	endpoint := i.client.projectPath("events", "batch-ingest")
	var out types.BatchIngestEventsResponse
	err := i.client.request(ctx, http.MethodPost, endpoint, nil, req, &out)
	return &out, err
}

func (i *IngestClient) BatchIngest(ctx context.Context, source string, events []BatchIngestEvent) (*types.BatchIngestEventsResponse, error) {
	requests := make([]types.BatchIngestEventRequest, 0, len(events))
	for _, event := range events {
		raw, err := BuildRawJSON(event.RawJSON, nil, event.PricingRuleSetID)
		if err != nil {
			return nil, err
		}
		requests = append(requests, types.BatchIngestEventRequest{
			ExtractorRuleSetID: event.ExtractorRuleSetID,
			Source:             event.Source,
			ExternalEventID:    event.ExternalEventID,
			IdempotencyKey:     event.IdempotencyKey,
			SubjectType:        event.SubjectType,
			SubjectID:          event.SubjectID,
			OccurredAt:         event.OccurredAt,
			RawJSON:            raw,
			XBilling:           event.XBilling,
			Metadata:           event.Metadata,
		})
	}
	return i.BatchIngestRawEvents(ctx, types.BatchIngestEventsRequest{
		Source: source,
		Events: requests,
	})
}

func BuildRawJSON(payload any, xBilling *types.XBilling, pricingRuleSetID string) (json.RawMessage, error) {
	raw, err := marshalRawObject(payload)
	if err != nil {
		return nil, err
	}
	if xBilling == nil && pricingRuleSetID == "" {
		return raw, nil
	}
	var object map[string]any
	if err = json.Unmarshal(raw, &object); err != nil {
		return nil, errors.New("raw json must be a JSON object when x-billing or x-pricing-rule is attached")
	}
	if xBilling != nil {
		object["x-billing"] = xBilling
	}
	if pricingRuleSetID != "" {
		object["x-pricing-rule"] = pricingRuleSetID
	}
	return json.Marshal(object)
}

func marshalRawObject(payload any) (json.RawMessage, error) {
	switch value := payload.(type) {
	case nil:
		return json.RawMessage(`{}`), nil
	case json.RawMessage:
		return append(json.RawMessage(nil), value...), nil
	case []byte:
		return append(json.RawMessage(nil), value...), nil
	case string:
		return json.RawMessage(value), nil
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}
}
