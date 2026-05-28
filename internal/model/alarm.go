package model

import "time"

// Alarm 告警记录模型，对应 alarms 表。
// 由 MQTT tdw/robot/{id}/alarm 消息自动创建，支持多级告警和处理流程。
type Alarm struct {
	AlarmID      string     `gorm:"primaryKey;type:varchar(32)" json:"alarm_id"` // 告警唯一标识
	AlarmLevel   int8       `gorm:"type:tinyint;not null" json:"alarm_level"`    // 告警级别: 1提示 2一般 3严重 4紧急
	AlarmType    string     `gorm:"type:varchar(50)" json:"alarm_type"`          // 告警类型编码（如 MOTOR_FAULT, SENSOR_ERROR）
	RobotID      string     `gorm:"type:varchar(32);index" json:"robot_id"`      // 关联机器人 ID
	StationID    string     `gorm:"type:varchar(32);index" json:"station_id"`    // 关联电站 ID（从机器人表自动获取）
	AlarmContent string     `gorm:"type:varchar(500)" json:"alarm_content"`      // 告警描述内容
	AlarmTime    time.Time  `gorm:"type:datetime" json:"alarm_time"`             // 告警产生时间（来自机器人时间戳）
	HandleStatus int8       `gorm:"type:tinyint;default:0" json:"handle_status"` // 处理状态: 0未处理 1已确认 2处理中 3已完成 4已忽略
	HandleTime   *time.Time `gorm:"type:datetime" json:"handle_time"`            // 处理时间
	HandlerID    string     `gorm:"type:varchar(32)" json:"handler_id"`          // 处理人 ID
	CreateTime   time.Time  `gorm:"autoCreateTime" json:"create_time"`           // 记录创建时间
}

func (Alarm) TableName() string {
	return "alarms"
}
