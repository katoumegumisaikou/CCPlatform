package model

import "time"

// NotifyTemplate 通知模板模型，对应 notify_templates 表。
// 定义不同告警级别/类型的通知内容模板和渠道配置。
type NotifyTemplate struct {
	TplID           string    `gorm:"primaryKey;type:varchar(32)" json:"tpl_id"`
	TplName         string    `gorm:"type:varchar(100);not null" json:"tpl_name"`
	TplType         string    `gorm:"type:varchar(20);index" json:"tpl_type"`       // 渠道类型: sms/email/app/webhook
	AlarmLevel      int8      `gorm:"type:tinyint" json:"alarm_level"`              // 适用告警级别，0 表示全部
	ChannelConfig   string    `gorm:"type:text" json:"channel_config"`              // 渠道配置 (JSON)
	ContentTemplate string    `gorm:"type:text" json:"content_template"`            // 内容模板，支持变量替换
	Status          int8      `gorm:"type:tinyint;default:1" json:"status"`         // 状态: 0禁用 1启用
	CreateTime      time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (NotifyTemplate) TableName() string {
	return "notify_templates"
}
