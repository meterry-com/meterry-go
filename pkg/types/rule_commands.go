package types

import "encoding/json"

type CreateExtractorRuleSetCommand struct {
	ID            string
	TenantID      string
	ProjectID     string
	Name          string
	TemplateID    string
	Version       int
	Status        RuleSetStatus
	EffectiveFrom int64
	Subject       SubjectExtractor
	Dimensions    []DimensionExtractor
	ChargeItems   []ChargeItemExtractor
}

type CreateExtractorRuleSetTemplateCommand struct {
	ID          string
	TenantID    string
	ProjectID   string
	Name        string
	Description string
	Fields      json.RawMessage
}

type UpdateExtractorRuleSetTemplateCommand struct {
	ID          string
	TenantID    string
	ProjectID   string
	Name        string
	Description string
	Fields      json.RawMessage
}

type CreatePricingRuleSetCommand struct {
	ID                 string
	TenantID           string
	ProjectID          string
	ExtractorRuleSetID string
	Name               string
	Version            int
	Status             RuleSetStatus
	Currency           string
	EffectiveFrom      int64
	Rules              []PricingRule
}

type TestRulesCommand struct {
	TenantID           string
	ProjectID          string
	ExtractorRuleSetID string
	PricingRuleSetID   string
	RawJSON            json.RawMessage
}
