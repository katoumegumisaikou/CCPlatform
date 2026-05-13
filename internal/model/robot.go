package model

import "time"

// Robot 光伏清扫机器人模型，对应 robots 表。
// 记录机器人的实时状态数据，由 MQTT 上行消息持续更新。
// 支持三种类型: 1固定式(轨道) 2接驳车 3全智能(轮式自主导航)。
type Robot struct {
	RobotID        string    `gorm:"primaryKey;type:varchar(32)" json:"robot_id"`              // 机器人唯一标识
	RobotCode      string    `gorm:"type:varchar(50);uniqueIndex" json:"robot_code"`           // 机器人编码（唯一）
	RobotName      string    `gorm:"type:varchar(100)" json:"robot_name"`                      // 机器人名称
	RobotType      int8      `gorm:"type:tinyint;not null" json:"robot_type"`                  // 类型: 1固定式 2接驳车 3全智能
	StationID      string    `gorm:"type:varchar(32);index" json:"station_id"`                 // 所属电站 ID
	PosX           float64   `gorm:"type:decimal(10,3)" json:"pos_x"`                          // X 坐标 (m)，由 MQTT position 消息更新
	PosY           float64   `gorm:"type:decimal(10,3)" json:"pos_y"`                          // Y 坐标 (m)
	PosZ           float64   `gorm:"type:decimal(10,3)" json:"pos_z"`                          // Z 坐标 (m)
	Heading        float64   `gorm:"type:decimal(5,2)" json:"heading"`                         // 朝向角度 (度)
	BatteryLevel   int8      `gorm:"type:tinyint" json:"battery_level"`                        // 电量百分比，由 MQTT heartbeat 消息更新
	OnlineStatus   int8      `gorm:"type:tinyint;default:0" json:"online_status"`              // 在线状态: 0离线 1在线
	WorkStatus     int8      `gorm:"type:tinyint;default:0" json:"work_status"`                // 工作状态: 0空闲 1清扫中 2充电中 3故障 4维护
	Speed          float64   `gorm:"type:decimal(5,2)" json:"speed"`                           // 速度 (m/s)
	CleanArea      float64   `gorm:"type:decimal(10,2)" json:"clean_area"`                     // 累计清扫面积 (㎡)
	FaultCode      int       `gorm:"type:int;default:0" json:"fault_code"`                     // 故障码，0 表示无故障
	Temperature    float64   `gorm:"type:decimal(5,2)" json:"temperature"`                     // 环境温度 (℃)
	Humidity       float64   `gorm:"type:decimal(5,2)" json:"humidity"`                        // 环境湿度 (%)
	LightIntensity float64   `gorm:"type:decimal(10,2)" json:"light_intensity"`                // 光照强度 (lux)
	WindSpeed      float64   `gorm:"type:decimal(5,2)" json:"wind_speed"`                      // 风速 (m/s)
	LastHeartbeat  time.Time `gorm:"type:datetime" json:"last_heartbeat"`                      // 最后心跳时间，用于判断在线状态
	CreateTime     time.Time `gorm:"autoCreateTime" json:"create_time"`                        // 注册时间
	UpdateTime     time.Time `gorm:"autoUpdateTime" json:"update_time"`                        // 最后更新时间
}

func (Robot) TableName() string {
	return "robots"
}
