package model

import "time"

// RobotConfig 设备配置模板模型，对应 robot_configs 表。
// 存储机器人工作参数配置模板，支持按机器人类型设置默认模板。
type RobotConfig struct {
	ConfigID            string    `gorm:"primaryKey;type:varchar(32)" json:"config_id"`
	ConfigName          string    `gorm:"type:varchar(100);not null" json:"config_name"`
	RobotType           int8      `gorm:"type:tinyint" json:"robot_type"`           // 适用机器人类型，0 表示通用
	ConfigData          string    `gorm:"type:text" json:"config_data"`             // 配置参数 (JSON)
	ScheduleConfig      string    `gorm:"type:text" json:"schedule_config"`         // 时段配置 JSON: 星期、3 个时间段、启用开关
	ControlConfig       string    `gorm:"type:text" json:"control_config"`          // 控制配置 JSON: 车辆类型、清扫方向、运行模式、安装位置
	ThresholdConfig     string    `gorm:"type:text" json:"threshold_config"`        // 阈值配置 JSON: 清扫车/接驳车报警电压、空转电流、过载电流
	ServerConfig        string    `gorm:"type:text" json:"server_config"`           // 服务器配置 JSON: IP、端口、时区
	CommunicationConfig string    `gorm:"type:text" json:"communication_config"`    // 通信参数 JSON: 4G/5G、LoRa、WiFi、星闪
	IsDefault           int8      `gorm:"type:tinyint;default:0" json:"is_default"` // 是否为默认模板
	Status              int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime          time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime          time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (RobotConfig) TableName() string {
	return "robot_configs"
}
