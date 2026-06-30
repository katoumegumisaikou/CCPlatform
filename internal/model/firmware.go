package model

import "time"

// Firmware 固件模型，对应 firmwares 表。
// 记录固件版本信息、升级包路径和兼容的机器人类型。
type Firmware struct {
	FirmwareID           string    `gorm:"primaryKey;type:varchar(32)" json:"firmware_id"`
	Version              string    `gorm:"type:varchar(50);not null" json:"version"`
	PackagePath          string    `gorm:"type:varchar(500)" json:"package_path"`
	CompatibleRobotTypes string    `gorm:"type:varchar(200)" json:"compatible_robot_types"` // 兼容机器人类型，逗号分隔
	UpgradeLog           string    `gorm:"type:text" json:"upgrade_log"`                    // 升级日志 (JSON)
	Status               int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime           time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime           time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (Firmware) TableName() string {
	return "firmwares"
}
