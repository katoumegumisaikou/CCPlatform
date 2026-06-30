package model

import "time"

// Geofence 电子围栏模型，对应 geofences 表。
// 支持圆形围栏（center_lng/center_lat/radius）和多边形围栏（points 顶点坐标）。
// 动作类型：禁止进入（机器人进入围栏区域触发告警）或禁止离开（机器人离开围栏区域触发告警）。
type Geofence struct {
	GeofenceID  string    `gorm:"primaryKey;type:varchar(32)" json:"geofence_id"`   // 围栏唯一标识
	FenceName   string    `gorm:"type:varchar(100);not null" json:"fence_name"`      // 围栏名称
	FenceType   int8      `gorm:"type:tinyint;not null" json:"fence_type"`           // 围栏形状: 1=圆形 2=多边形
	ActionType  int8      `gorm:"type:tinyint;not null" json:"action_type"`          // 动作类型: 1=禁止进入 2=禁止离开
	ScopeType   string    `gorm:"type:varchar(20)" json:"scope_type"`                // 作用域: global/station/robot
	ScopeID     string    `gorm:"type:varchar(32);index" json:"scope_id"`             // 作用域 ID（电站 ID 或机器人 ID）
	CenterLng   float64   `gorm:"type:decimal(10,7)" json:"center_lng"`              // 圆心经度（圆形围栏必填）
	CenterLat   float64   `gorm:"type:decimal(10,7)" json:"center_lat"`              // 圆心纬度（圆形围栏必填）
	Radius      float64   `gorm:"type:decimal(10,2)" json:"radius"`                  // 半径（圆形围栏必填，单位米）
	Points      string    `gorm:"type:text" json:"points"`                           // 多边形顶点坐标 JSON: [[lng,lat],[lng,lat],...]
	AlarmLevel  int8      `gorm:"type:tinyint;default:2" json:"alarm_level"`         // 触发告警级别: 1提示 2一般 3严重 4紧急
	Description string    `gorm:"type:varchar(500)" json:"description"`              // 围栏描述
	Color       string    `gorm:"type:varchar(20);default:'#ff4d4f'" json:"color"`   // 地图显示颜色
	Status      int8      `gorm:"type:tinyint;default:1" json:"status"`              // 状态: 0禁用 1启用
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`                 // 创建时间
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`                 // 更新时间
}

func (Geofence) TableName() string {
	return "geofences"
}

// GeofenceAlarm 围栏告警记录模型，对应 geofence_alarms 表。
// 当机器人触发电子围栏规则时自动创建记录，同时推送到告警中心。
type GeofenceAlarm struct {
	AlarmID      string     `gorm:"primaryKey;type:varchar(32)" json:"alarm_id"`       // 告警唯一标识
	FenceID      string     `gorm:"type:varchar(32);index" json:"fence_id"`            // 关联围栏 ID
	FenceName    string     `gorm:"type:varchar(100)" json:"fence_name"`               // 围栏名称（冗余，便于展示）
	RobotID      string     `gorm:"type:varchar(32);index" json:"robot_id"`            // 触发机器人 ID
	RobotName    string     `gorm:"type:varchar(100)" json:"robot_name"`               // 机器人名称（冗余）
	StationID    string     `gorm:"type:varchar(32);index" json:"station_id"`          // 所属电站 ID
	TriggerType  int8       `gorm:"type:tinyint" json:"trigger_type"`                  // 触发类型: 1=进入禁区 2=离开工作区
	PositionLng  float64    `gorm:"type:decimal(10,7)" json:"position_lng"`            // 触发时经度
	PositionLat  float64    `gorm:"type:decimal(10,7)" json:"position_lat"`            // 触发时纬度
	AlarmContent string     `gorm:"type:varchar(500)" json:"alarm_content"`            // 告警内容描述
	AlarmTime    time.Time  `gorm:"type:datetime" json:"alarm_time"`                   // 触发时间
	HandleStatus int8       `gorm:"type:tinyint;default:0" json:"handle_status"`       // 处理状态: 0未处理 1已确认 2处理中 3已完成 4已忽略
	HandleTime   *time.Time `gorm:"type:datetime" json:"handle_time"`                  // 处理时间
	HandleRemark string     `gorm:"type:varchar(500)" json:"handle_remark"`            // 处理备注
	CreateTime   time.Time  `gorm:"autoCreateTime" json:"create_time"`                 // 记录创建时间
}

func (GeofenceAlarm) TableName() string {
	return "geofence_alarms"
}
