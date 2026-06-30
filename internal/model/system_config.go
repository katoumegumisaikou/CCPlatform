package model

import "time"

// SystemConfig 系统配置模型，对应 system_configs 表。
// 存储平台全局参数，按 category 分组管理。
type SystemConfig struct {
	ConfigKey   string    `gorm:"primaryKey;type:varchar(100)" json:"config_key"`
	ConfigValue string    `gorm:"type:text" json:"config_value"`
	Category    string    `gorm:"type:varchar(50);index" json:"category"` // 配置分类: notify/ota/backup/video/general
	Description string    `gorm:"type:varchar(200)" json:"description"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
