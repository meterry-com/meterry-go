package types

const (
	APP_NAME = "billing"
)

// Watermark 水位线表结构
type Watermark struct {
	ID                 int64 `json:"id" db:"id"`                                     // 主键ID，通常固定为1，因为只需要一条记录
	CurrentWatermark   int64 `json:"current_watermark" db:"current_watermark"`       // 当前水位线位置
	ReconcilePosition  int64 `json:"reconcile_position" db:"reconcile_position"`     // 当前账单校准位置
	WatermarkUpdatedAt int64 `json:"watermark_updated_at" db:"watermark_updated_at"` // 水位线更新时间(unix timestamp)
	ReconcileUpdatedAt int64 `json:"reconcile_updated_at" db:"reconcile_updated_at"` // 校准水位线更新时间(unix timestamp)
}

const (
	// 	根据etcd官方GitHub仓库中的一个相关问题讨论（Issue #11187），etcd团队成员表示不需要担心版本号溢出的问题 1 ：
	// "I will not worry about this... Even if you change the state of the store 20k every seconds, it requires 5.3 billion years to overflow it."
	// 这意味着即使每秒进行20,000次状态更改，也需要约53亿年才能达到int64的上限。这在实际应用中几乎是不可能发生的情况。
	MAX_BALANCE_VERSION = 9223372036854770000
)
