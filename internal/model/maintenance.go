package model

import "time"

// Maintenance 维护记录模型，对应 maintenance 表。
// 记录机器人的维护历史，包括故障描述和配件更换信息。
type Maintenance struct {
	MaintID       string    `gorm:"primaryKey;type:varchar(32)" json:"maint_id"`
	RobotID       string    `gorm:"type:varchar(32);index" json:"robot_id"`
	StationID     string    `gorm:"type:varchar(32);index" json:"station_id"`
	MaintType     string    `gorm:"type:varchar(50)" json:"maint_type"`     // 维护类型: routine/repair/overhaul
	Handler       string    `gorm:"type:varchar(50)" json:"handler"`        // 处理人
	FaultDesc     string    `gorm:"type:varchar(500)" json:"fault_desc"`    // 故障描述
	PartsReplaced string    `gorm:"type:varchar(500)" json:"parts_replaced"` // 更换配件
	MaintTime     *time.Time `gorm:"type:datetime" json:"maint_time"`        // 维护时间
	NextMaintTime *time.Time `gorm:"type:datetime" json:"next_maint_time"`  // 下次维护时间（用于提醒）
	CreateTime    time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime    time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (Maintenance) TableName() string {
	return "maintenance"
}
