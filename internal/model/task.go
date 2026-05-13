package model

import "time"

// Task 清扫任务模型，对应 tasks 表。
// 记录清扫任务的计划和执行信息，支持即时/定时/周期三种任务类型。
// 任务状态流转: 待执行(0) → 执行中(1) → 已完成(2)
//                         → 已暂停(3) → 已取消(4)
//                         → 失败(5)
type Task struct {
	TaskID      string     `gorm:"primaryKey;type:varchar(32)" json:"task_id"`      // 任务唯一标识
	TaskName    string     `gorm:"type:varchar(100)" json:"task_name"`              // 任务名称
	TaskType    int8       `gorm:"type:tinyint;not null" json:"task_type"`          // 任务类型: 1即时 2定时 3周期
	RobotID     string     `gorm:"type:varchar(32);index" json:"robot_id"`          // 执行机器人 ID
	StationID   string     `gorm:"type:varchar(32);index" json:"station_id"`        // 所属电站 ID（创建时自动从机器人获取）
	AreaIDs     string     `gorm:"type:varchar(500)" json:"area_ids"`               // 清扫区域 ID 列表（逗号分隔）
	PlanStart   *time.Time `gorm:"type:datetime" json:"plan_start"`                 // 计划开始时间
	PlanEnd     *time.Time `gorm:"type:datetime" json:"plan_end"`                   // 计划结束时间
	ActualStart *time.Time `gorm:"type:datetime" json:"actual_start"`               // 实际开始时间
	ActualEnd   *time.Time `gorm:"type:datetime" json:"actual_end"`                 // 实际结束时间
	CleanArea   float64    `gorm:"type:decimal(10,2)" json:"clean_area"`            // 实际清扫面积 (㎡)
	TaskStatus  int8       `gorm:"type:tinyint;default:0" json:"task_status"`       // 任务状态: 0待执行 1执行中 2已完成 3已暂停 4已取消 5失败
	CreateTime  time.Time  `gorm:"autoCreateTime" json:"create_time"`               // 创建时间
	UpdateTime  time.Time  `gorm:"autoUpdateTime" json:"update_time"`               // 更新时间
}

func (Task) TableName() string {
	return "tasks"
}
