package types

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/shopspring/decimal"
)

type RawEvent struct {
	ID                 string            `json:"id"`                              // 原始事件ID
	TenantID           string            `json:"tenant_id"`                       // 租户ID
	ProjectID          string            `json:"project_id,omitempty"`            // Project 是接入和 usage analytics 配置边界
	Source             string            `json:"source"`                          // 事件来源，例如 openai-gateway
	ExternalEventID    string            `json:"external_event_id,omitempty"`     // 外部系统事件ID
	IdempotencyKey     string            `json:"idempotency_key,omitempty"`       // 幂等键，用于防止重复入库
	ExtractorRuleSetID string            `json:"extractor_rule_set_id,omitempty"` // 入站指定的提取规则集ID
	SubjectType        string            `json:"subject_type,omitempty"`          // 入站显式指定的主对象类型，优先于提取规则
	SubjectID          string            `json:"subject_id,omitempty"`            // 入站显式指定的主对象ID，必须与subject_type成对出现
	OccurredAt         int64             `json:"occurred_at"`                     // 事件实际发生时间戳
	RawJSON            json.RawMessage   `json:"raw_json"`                        // 原始用量JSON，保持字节形式用于高吞吐提取
	IngestedAt         int64             `json:"ingested_at"`                     // 平台接收时间戳
	Status             RawEventStatus    `json:"status"`                          // 接收状态，处理状态记录在 processing log
	ErrorMessage       string            `json:"error_message,omitempty"`         // 接收阶段错误信息
	Metadata           map[string]string `json:"metadata,omitempty"`              // 入站元数据，例如IP、网关标记
}

type RawEventStatus string

const (
	RawEventStatusReceived RawEventStatus = "received"
	RawEventStatusParsed   RawEventStatus = "parsed"
	RawEventStatusFailed   RawEventStatus = "failed"
)

type RawEventProcessingLog struct {
	ID           string         `json:"id" db:"id"`                                 // 处理日志ID
	RawEventID   string         `json:"raw_event_id" db:"raw_event_id"`             // 对应的原始事件ID
	TenantID     string         `json:"tenant_id" db:"tenant_id"`                   // 租户ID
	ProjectID    string         `json:"project_id,omitempty" db:"project_id"`       // Project 归属
	Consumer     string         `json:"consumer" db:"consumer"`                     // 消费者名称或consumer group
	Status       RawEventStatus `json:"status" db:"status"`                         // 本次处理结果
	ErrorMessage string         `json:"error_message,omitempty" db:"error_message"` // 本次处理错误信息
	ProcessedAt  int64          `json:"processed_at" db:"processed_at"`             // 处理完成时间戳
}

type UsageEvent struct {
	ID                  string            `json:"id"`                           // 标准化用量事件ID
	TenantID            string            `json:"tenant_id"`                    // 租户ID
	ProjectID           string            `json:"project_id,omitempty"`         // Project 是接入和 usage analytics 配置边界
	RawEventID          string            `json:"raw_event_id"`                 // 来源原始事件ID
	Source              string            `json:"source"`                       // 事件来源
	ExternalEventID     string            `json:"external_event_id,omitempty"`  // 外部系统事件ID
	SubjectType         string            `json:"subject_type,omitempty"`       // 租户内部主查询/计费对象类型，例如user、workspace
	SubjectID           string            `json:"subject_id,omitempty"`         // 租户内部主查询/计费对象ID
	BillingAccountID    string            `json:"billing_account_id,omitempty"` // route解析后的真实扣费账户快照
	BillingRouteID      string            `json:"billing_route_id,omitempty"`   // route版本快照标识
	BillingRouteVersion int               `json:"billing_route_version,omitempty"`
	BudgetLeaseID       string            `json:"budget_lease_id,omitempty"` // 一次性 agent run 预算租约快照
	ExtractorRuleSetID  string            `json:"extractor_rule_set_id"`     // 使用的提取规则集ID
	ExtractorVersion    int               `json:"extractor_version"`         // 使用的提取规则集版本
	OccurredAt          int64             `json:"occurred_at"`               // 事件发生时间戳
	Dimensions          map[string]string `json:"dimensions"`                // 标准化维度，例如provider、model、project_id
	ChargeItems         []ChargeItem      `json:"charge_items,omitempty"`    // rating 后内嵌保存的计费事实
	NormalizeStatus     NormalizeStatus   `json:"normalize_status"`          // 标准化处理状态
	ErrorMessage        string            `json:"error_message,omitempty"`   // 标准化错误信息
	Version             int64             `json:"version,omitempty"`         // 处理批次水位版本，用于高吞吐账单校准
	CreatedAt           int64             `json:"created_at"`                // 创建时间戳
}

type NormalizeStatus string

const (
	NormalizeStatusSuccess NormalizeStatus = "success"
	NormalizeStatusFailed  NormalizeStatus = "failed"
)

type ChargeItem struct {
	ID               string            `json:"id"`                            // 计量项ID，基于原始事件和提取规则稳定生成
	UsageEventID     string            `json:"usage_event_id"`                // 所属标准化用量事件ID
	SubjectType      string            `json:"subject_type,omitempty"`        // 租户内部主查询/计费对象类型
	Metric           string            `json:"metric"`                        // 计量指标，例如prompt_tokens
	Quantity         float64           `json:"quantity"`                      // 计量数量
	Unit             string            `json:"unit"`                          // 计量单位，例如token、request
	Attributes       map[string]string `json:"attributes,omitempty"`          // 计量项附加属性
	Control          map[string]string `json:"control,omitempty"`             // 计费控制字段，例如是否允许使用代金券
	PricingRuleSetID string            `json:"pricing_rule_set_id,omitempty"` // 使用的定价规则集ID
	PricingVersion   int               `json:"pricing_version,omitempty"`     // 使用的定价规则集版本
	MatchedRuleID    string            `json:"matched_rule_id,omitempty"`     // 命中的定价审计键，fallback pricing 使用metric
	Currency         string            `json:"currency,omitempty"`            // 币种
	UnitPrice        float64           `json:"unit_price,omitempty"`          // 单价
	Amount           float64           `json:"amount,omitempty"`              // 计费金额
	Version          int64             `json:"version,omitempty"`             // 处理批次水位版本，用于高吞吐账单校准
}

type UsageEventWithChargeItems struct {
	UsageEvent  *UsageEvent  `json:"usage_event"`
	ChargeItems []ChargeItem `json:"charge_items"`
}

type UsageLogMessage struct {
	TenantID           string                 `json:"tenant_id"`
	ProjectID          string                 `json:"project_id"`
	EventDate          string                 `json:"event_date"`
	OccurredAt         int64                  `json:"occurred_at"`
	UsageEventID       string                 `json:"usage_event_id"`
	RawEventID         string                 `json:"raw_event_id"`
	Source             string                 `json:"source"`
	ExternalEventID    string                 `json:"external_event_id,omitempty"`
	SubjectType        string                 `json:"subject_type,omitempty"`
	SubjectID          string                 `json:"subject_id,omitempty"`
	BillingAccountID   string                 `json:"billing_account_id,omitempty"`
	ExtractorRuleSetID string                 `json:"extractor_rule_set_id,omitempty"`
	ExtractorVersion   int                    `json:"extractor_version,omitempty"`
	LabelsJSON         map[string]string      `json:"labels_json,omitempty"`
	MetricsJSON        map[string]UsageMetric `json:"metrics_json,omitempty"`
	LogJSON            map[string]any         `json:"log_json,omitempty"`
	UsageVersion       int64                  `json:"usage_version,omitempty"`
	CreatedAt          int64                  `json:"created_at,omitempty"`
}

type UsageLogPublishStatus string

const (
	UsageLogPublishStatusPending   UsageLogPublishStatus = "pending"
	UsageLogPublishStatusPublished UsageLogPublishStatus = "published"
	UsageLogPublishStatusFailed    UsageLogPublishStatus = "failed"
)

type UsageLogPublishOutbox struct {
	TenantID     string                `json:"tenant_id"`
	ProjectID    string                `json:"project_id"`
	RawEventID   string                `json:"raw_event_id"`
	UsageEventID string                `json:"usage_event_id"`
	Status       UsageLogPublishStatus `json:"status"`
	AttemptCount int                   `json:"attempt_count"`
	LastError    string                `json:"last_error,omitempty"`
	CreatedAt    int64                 `json:"created_at"`
	UpdatedAt    int64                 `json:"updated_at"`
	PublishedAt  int64                 `json:"published_at,omitempty"`
	UsageVersion int64                 `json:"usage_version"`
}

type ChargeItemAmountAggregate struct {
	TenantID         string          `json:"tenant_id"`
	BillingAccountID string          `json:"billing_account_id,omitempty"`
	SubjectType      string          `json:"subject_type,omitempty"`
	SubjectID        string          `json:"subject_id,omitempty"`
	Currency         string          `json:"currency"`
	Amount           decimal.Decimal `json:"amount"`
	MaxVersion       int64           `json:"max_version"`
	ItemCount        int             `json:"item_count"`
}

// UsageChargeDetailKind identifies the runtime projection carried by a
// UsageChargeDetail. It is not part of the durable or public charge detail
// contract: a rated charge item can contribute to both projections.
type UsageChargeDetailKind string

const (
	UsageChargeDetailKindQuantity UsageChargeDetailKind = "quantity"
	UsageChargeDetailKindAmount   UsageChargeDetailKind = "amount"
)

type UsageChargeDetail struct {
	ID               string            `json:"id"`
	TenantID         string            `json:"tenant_id"`
	ProjectID        string            `json:"project_id,omitempty"`
	UsageEventID     string            `json:"usage_event_id"`
	RawEventID       string            `json:"raw_event_id"`
	SubjectType      string            `json:"subject_type,omitempty"`
	SubjectID        string            `json:"subject_id,omitempty"`
	BillingAccountID string            `json:"billing_account_id,omitempty"`
	BudgetLeaseID    string            `json:"budget_lease_id,omitempty"`
	Currency         string            `json:"currency"`
	Amount           decimal.Decimal   `json:"amount"`
	Control          map[string]string `json:"control,omitempty"`
	// Kind, LimitDimensions, and LimitItems are transient post-persist inputs for the
	// realtime usage-control projection. They are intentionally excluded from
	// public and durable DTOs.
	Kind                    UsageChargeDetailKind      `json:"-"`
	LimitDimensions         map[string]string          `json:"-"`
	LimitItems              []ChargeItem               `json:"-"`
	LimitUsageByMetricScope map[string]decimal.Decimal `json:"limit_usage_by_metric_scope,omitempty"`
	LimitQuantityBySelector map[string]decimal.Decimal `json:"limit_quantity_by_selector,omitempty"`
	LimitAmountByMetric     map[string]decimal.Decimal `json:"limit_amount_by_metric,omitempty"`
	ChargeItemCount         int                        `json:"charge_item_count"`
	UsageVersion            int64                      `json:"usage_version,omitempty"`
	OccurredAt              int64                      `json:"occurred_at"`
	CreatedAt               int64                      `json:"created_at"`
}

type UsageFundDebitDetail struct {
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	ProjectID        string          `json:"project_id,omitempty"`
	UsageEventID     string          `json:"usage_event_id"`
	RawEventID       string          `json:"raw_event_id"`
	SubjectType      string          `json:"subject_type,omitempty"`
	SubjectID        string          `json:"subject_id,omitempty"`
	BillingAccountID string          `json:"billing_account_id"`
	Currency         string          `json:"currency"`
	FundType         FundType        `json:"fund_type"`
	FundID           string          `json:"fund_id"`
	Amount           decimal.Decimal `json:"amount"`
	UsageVersion     int64           `json:"usage_version,omitempty"`
	CreatedAt        int64           `json:"created_at"`
}

type FundDebitAggregate struct {
	TenantID         string          `json:"tenant_id,omitempty"`
	ProjectID        string          `json:"project_id,omitempty"`
	BillingAccountID string          `json:"billing_account_id,omitempty"`
	Currency         string          `json:"currency,omitempty"`
	FundType         FundType        `json:"fund_type,omitempty"`
	FundID           string          `json:"fund_id,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	MaxVersion       int64           `json:"max_version"`
	ItemCount        int             `json:"item_count"`
}

type UsageReconcileAccountAggregate struct {
	TenantID          string `json:"tenant_id"`
	ProjectID         string `json:"project_id"`
	AccountID         string `json:"account_id"`
	MaxVersion        int64  `json:"max_version"`
	FundDebitRowCount int    `json:"fund_debit_row_count"`
}

type UsageInterimBill struct {
	ID                string          `json:"id"`
	TenantID          string          `json:"tenant_id"`
	ProjectID         string          `json:"project_id"`
	AccountID         string          `json:"account_id"`
	Currency          string          `json:"currency"`
	FundType          FundType        `json:"fund_type"`
	FundID            string          `json:"fund_id"`
	Amount            decimal.Decimal `json:"amount"`
	ChargeItemCount   int             `json:"charge_item_count"`
	StartUsageVersion int64           `json:"start_usage_version"`
	EndUsageVersion   int64           `json:"end_usage_version"`
	IdempotencyKey    string          `json:"idempotency_key"`
	CreatedAt         int64           `json:"created_at"`
}

type UsageReconcileAccountCommand struct {
	TenantID          string `json:"tenant_id"`
	ProjectID         string `json:"project_id"`
	AccountID         string `json:"account_id"`
	StartUsageVersion int64  `json:"start_usage_version,omitempty"`
	EndUsageVersion   int64  `json:"end_usage_version,omitempty"`
	IdempotencyKey    string `json:"idempotency_key"`
	CreatedAt         int64  `json:"created_at,omitempty"`
}

type UsageReconcileState struct {
	TenantID          string `json:"tenant_id"`
	AccountID         string `json:"account_id"`
	ReconciledVersion int64  `json:"reconciled_usage_version"`
	UpdatedAt         int64  `json:"updated_at"`
}

type UsageLogRow struct {
	TenantID         string            `json:"tenant_id"`
	ProjectID        string            `json:"project_id"`
	BillingAccountID string            `json:"-"`
	ChargeItemID     string            `json:"charge_item_id"`
	UsageEventID     string            `json:"usage_event_id"`
	RawEventID       string            `json:"raw_event_id"`
	OccurredAt       int64             `json:"occurred_at"`
	SubjectType      string            `json:"subject_type,omitempty"`
	SubjectID        string            `json:"subject_id,omitempty"`
	Metric           string            `json:"metric"`
	Quantity         decimal.Decimal   `json:"quantity"`
	Unit             string            `json:"unit"`
	Currency         string            `json:"currency,omitempty"`
	Amount           decimal.Decimal   `json:"amount"`
	Labels           map[string]string `json:"labels,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
	PricingRuleSetID string            `json:"pricing_rule_set_id,omitempty"`
	PricingVersion   int               `json:"pricing_version,omitempty"`
	MatchedRuleID    string            `json:"matched_rule_id,omitempty"`
	UsageVersion     int64             `json:"usage_version,omitempty"`
	CreatedAt        int64             `json:"created_at"`
}

type UsageMetric struct {
	Value         string `json:"value"`
	Unit          string `json:"unit,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Amount        string `json:"amount,omitempty"`
	UnitPrice     string `json:"unit_price,omitempty"`
	MatchedRuleID string `json:"matched_rule_id,omitempty"`
}

// OrderedUsageMetrics is serialized in metric-name order so usage detail
// responses are stable for API consumers and display consistently in the portal.
type OrderedUsageMetrics map[string]UsageMetric

func (m OrderedUsageMetrics) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			buf.WriteByte(',')
		}
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		encodedValue, err := json.Marshal(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(encodedKey)
		buf.WriteByte(':')
		buf.Write(encodedValue)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type UsageEventLogRow struct {
	TenantID         string              `json:"tenant_id"`
	ProjectID        string              `json:"project_id"`
	EventDate        string              `json:"event_date"`
	OccurredAt       int64               `json:"occurred_at"`
	UsageEventID     string              `json:"usage_event_id"`
	RawEventID       string              `json:"raw_event_id"`
	Source           string              `json:"source"`
	ExternalEventID  string              `json:"external_event_id,omitempty"`
	SubjectType      string              `json:"subject_type,omitempty"`
	SubjectID        string              `json:"subject_id,omitempty"`
	BillingAccountID string              `json:"billing_account_id,omitempty"`
	Labels           map[string]string   `json:"labels,omitempty"`
	Metrics          OrderedUsageMetrics `json:"metrics,omitempty"`
	Log              map[string]any      `json:"log,omitempty"`
	UsageVersion     int64               `json:"usage_version,omitempty"`
	CreatedAt        int64               `json:"created_at"`
}

type XBilling struct {
	Items         []ReportedChargeItem         `json:"items,omitempty"`          // 由调用方直接上报的动态计费项
	PricingHints  map[string]ReportedPriceHint `json:"pricing_hints,omitempty"`  // 按metric提供的动态价格提示
	DiscountHints map[string]ReportedDiscount  `json:"discount_hints,omitempty"` // 按metric提供的动态折扣提示
}

type ReportedChargeItem struct {
	Metric     string            `json:"metric"`               // 计量指标，例如tool_call.web_search
	Quantity   float64           `json:"quantity"`             // 计量数量
	Unit       string            `json:"unit"`                 // 计量单位，例如call、minute
	Attributes map[string]string `json:"attributes,omitempty"` // 动态计费项附加属性
}

type ReportedPriceHint struct {
	UnitPrice   float64 `json:"unit_price"`             // 单价
	PricingUnit float64 `json:"pricing_unit,omitempty"` // 计价单位，例如1或1000，默认1
	Currency    string  `json:"currency,omitempty"`     // 币种，默认USD
}

type ReportedDiscount struct {
	Type   DiscountType `json:"type"`             // 折扣类型：percent或amount
	Value  float64      `json:"value"`            // 折扣值
	Reason string       `json:"reason,omitempty"` // 折扣原因
}

type DiscountType string

const (
	DiscountTypePercent DiscountType = "percent"
	DiscountTypeAmount  DiscountType = "amount"
)

type ExtractorRuleSet struct {
	ID             string                `json:"id"`                        // 提取规则集ID
	TenantID       string                `json:"tenant_id"`                 // 租户ID
	ProjectID      string                `json:"project_id,omitempty"`      // Project 归属，空值兼容租户级规则
	Name           string                `json:"name,omitempty"`            // 规则集名称，便于后台管理
	TemplateID     string                `json:"template_id,omitempty"`     // 关联的提取规则模板ID
	Version        int                   `json:"version"`                   // 规则集版本
	Status         RuleSetStatus         `json:"status"`                    // 规则集状态
	EffectiveFrom  int64                 `json:"effective_from"`            // 生效时间戳
	Subject        SubjectExtractor      `json:"subject,omitempty"`         // 主查询/计费对象提取规则
	Dimensions     []DimensionExtractor  `json:"dimensions"`                // 维度提取规则
	ChargeItems    []ChargeItemExtractor `json:"charge_item_extractors"`    // 计量项提取规则
	CreatedAt      int64                 `json:"created_at"`                // 创建时间戳
	ReadonlyFields []string              `json:"readonly_fields,omitempty"` // 模板填充且前端应锁定的字段路径

	TemplateFingerprint string `json:"-"` // 参与运行时缓存隔离的模板内容指纹
}

type ExtractorRuleSetTemplate struct {
	ID             string          `json:"id"`                        // 模板ID
	TenantID       string          `json:"tenant_id"`                 // 租户ID
	ProjectID      string          `json:"project_id,omitempty"`      // Project 归属，空值表示租户级模板
	Name           string          `json:"name"`                      // 模板名称
	Description    string          `json:"description,omitempty"`     // 模板描述
	Fields         json.RawMessage `json:"fields"`                    // ExtractorRuleSet字段patch
	ReadonlyFields []string        `json:"readonly_fields,omitempty"` // fields中会锁定的字段路径
	CreatedAt      int64           `json:"created_at"`                // 创建时间戳
	UpdatedAt      int64           `json:"updated_at"`                // 更新时间戳
}

type SubjectExtractor struct {
	DefaultType string `json:"default_type,omitempty"` // 缺失时默认主对象类型
	TypePath    string `json:"type_path,omitempty"`    // 主对象类型提取路径
	DefaultID   string `json:"default_id,omitempty"`   // 缺失时默认主对象ID
	IDPath      string `json:"id_path,omitempty"`      // 主对象ID提取路径
	Required    bool   `json:"required,omitempty"`     // 是否必须提取到完整主对象
}

type PricingRuleSet struct {
	ID                 string        `json:"id"`                    // 定价规则集ID
	TenantID           string        `json:"tenant_id"`             // 租户ID
	ProjectID          string        `json:"project_id,omitempty"`  // Project 归属，空值兼容租户级规则
	ExtractorRuleSetID string        `json:"extractor_rule_set_id"` // 绑定的提取规则集ID
	Name               string        `json:"name,omitempty"`        // 规则集名称，便于后台管理
	Version            int           `json:"version"`               // 规则集版本
	Status             RuleSetStatus `json:"status"`                // 规则集状态
	Currency           string        `json:"currency"`              // 默认计费币种
	EffectiveFrom      int64         `json:"effective_from"`        // 生效时间戳
	Rules              []PricingRule `json:"rules"`                 // 定价规则列表
	CreatedAt          int64         `json:"created_at"`            // 创建时间戳
}

type RuleSetStatus string

const (
	RuleSetStatusDraft     RuleSetStatus = "draft"
	RuleSetStatusPublished RuleSetStatus = "published"
	RuleSetStatusArchived  RuleSetStatus = "archived"
)

type DimensionExtractor struct {
	Name          string `json:"name"`                     // 维度名称
	Path          string `json:"path,omitempty"`           // 原始JSON提取路径
	Value         string `json:"value,omitempty"`          // 常量维度值
	Type          string `json:"type"`                     // 值类型，例如string
	Required      bool   `json:"required,omitempty"`       // 是否必填
	DefaultValue  string `json:"default_value,omitempty"`  // 缺失时默认值
	TransformExpr string `json:"transform_expr,omitempty"` // 预留转换表达式
}

type ChargeItemExtractor struct {
	Name          string            `json:"name"`                      // 提取器名称
	Metric        string            `json:"metric,omitempty"`          // 生成的计量指标
	Path          string            `json:"path,omitempty"`            // 原始JSON提取路径
	QuantityExpr  string            `json:"quantity_expr,omitempty"`   // 计量数量表达式，返回decimal字符串
	Type          string            `json:"type"`                      // 值类型，例如number
	Unit          string            `json:"unit"`                      // 计量单位
	SkipIfMissing bool              `json:"skip_if_missing,omitempty"` // 路径缺失时是否跳过
	Required      bool              `json:"required,omitempty"`        // 是否必填
	Attributes    map[string]string `json:"attributes,omitempty"`      // 生成计量项时附加的属性
	Control       map[string]string `json:"control,omitempty"`         // 生成计量项时附加的计费控制字段
	Currency      string            `json:"currency,omitempty"`        // attached pricing币种，默认USD
	AmountExpr    string            `json:"amount_expr,omitempty"`     // attached pricing金额表达式，返回decimal字符串
	UnitPrice     string            `json:"unit_price_expr,omitempty"` // attached pricing单价表达式，返回decimal字符串
}

type PricingRule struct {
	Name       string `json:"name"`                      // 定价规则名称
	Priority   int    `json:"priority"`                  // 匹配优先级，数值越小越先匹配
	Metric     string `json:"metric,omitempty"`          // 可选metric索引键，配置后仅作为该metric的候选规则
	MatchExpr  string `json:"match_expr"`                // CEL匹配表达式，返回bool
	AmountExpr string `json:"amount_expr"`               // CEL金额表达式，返回数值
	UnitPrice  string `json:"unit_price_expr,omitempty"` // CEL单价表达式，返回数值
}
