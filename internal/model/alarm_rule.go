package model

import "time"

// AlarmRule 告警规则模型，对应 alarm_rules 表。
// 定义告警的阈值触发条件、升级规则和抑制规则。
type AlarmRule struct {
	RuleID          string    `gorm:"primaryKey;type:varchar(32)" json:"rule_id"`
	RuleName        string    `gorm:"type:varchar(100);not null" json:"rule_name"`
	AlarmType       string    `gorm:"type:varchar(50);index" json:"alarm_type"`     // 适用的告警类型
	ThresholdValue  string    `gorm:"type:text" json:"threshold_value"`             // 阈值配置 (JSON)
	EscalationRule  string    `gorm:"type:text" json:"escalation_rule"`             // 升级规则 (JSON): [{count:3, within_min:10, to_level:4}]
	SuppressionRule string    `gorm:"type:text" json:"suppression_rule"`            // 抑制规则 (JSON): {within_sec:300}
	Status          int8      `gorm:"type:tinyint;default:1" json:"status"`         // 状态: 0禁用 1启用
	CreateTime      time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (AlarmRule) TableName() string {
	return "alarm_rules"
}
