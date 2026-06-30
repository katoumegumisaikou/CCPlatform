package model

import "time"

// RobotPosition 机器人位置历史记录模型，对应 robot_positions 表。
// 每次 MQTT position 消息上报时写入一条记录，用于轨迹回放。
type RobotPosition struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RobotID   string    `gorm:"type:varchar(32);index" json:"robot_id"`
	PosX      float64   `gorm:"type:decimal(10,3)" json:"pos_x"`
	PosY      float64   `gorm:"type:decimal(10,3)" json:"pos_y"`
	PosZ      float64   `gorm:"type:decimal(10,3)" json:"pos_z"`
	Heading   float64   `gorm:"type:decimal(5,2)" json:"heading"`
	Timestamp time.Time `gorm:"type:datetime;index" json:"timestamp"`
}

func (RobotPosition) TableName() string {
	return "robot_positions"
}
