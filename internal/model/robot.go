package model

import "time"

// Robot 光伏清扫机器人模型，对应 robots 表。
// 记录机器人的实时状态数据，由 MQTT 上行消息持续更新。
// 支持三种类型: 1固定式(轨道) 2接驳车 3全智能(轮式自主导航)。
type Robot struct {
	RobotID                 string     `gorm:"primaryKey;type:varchar(32)" json:"robot_id"`              // 机器人唯一标识
	RobotCode               string     `gorm:"type:varchar(50);uniqueIndex" json:"robot_code"`           // 机器人编码（唯一）
	RobotName               string     `gorm:"type:varchar(100)" json:"robot_name"`                      // 机器人名称
	RobotType               int8       `gorm:"type:tinyint;not null" json:"robot_type"`                  // 类型: 1固定式 2接驳车 3全智能
	StationID               string     `gorm:"type:varchar(32);index" json:"station_id"`                 // 所属电站 ID
	StationName             string     `gorm:"-" json:"station_name"`                                    // 所属电站名称（非持久化，Service 层填充）
	PosX                    float64    `gorm:"type:decimal(10,3)" json:"pos_x"`                          // X 坐标 (m)，由 MQTT position 消息更新
	PosY                    float64    `gorm:"type:decimal(10,3)" json:"pos_y"`                          // Y 坐标 (m)
	PosZ                    float64    `gorm:"type:decimal(10,3)" json:"pos_z"`                          // Z 坐标 (m)
	Heading                 float64    `gorm:"type:decimal(5,2)" json:"heading"`                         // 朝向角度 (度)
	GPSLongitude            float64    `gorm:"type:decimal(10,7)" json:"gps_longitude"`                  // 高精度 GPS 经度
	GPSLatitude             float64    `gorm:"type:decimal(10,7)" json:"gps_latitude"`                   // 高精度 GPS 纬度
	GPSAltitude             float64    `gorm:"type:decimal(10,3)" json:"gps_altitude"`                   // GPS 海拔/高度 (m)
	GPSAccuracy             float64    `gorm:"type:decimal(6,3)" json:"gps_accuracy"`                    // GPS 定位精度 (m)
	BatteryLevel            int8       `gorm:"type:tinyint" json:"battery_level"`                        // 电量百分比，由 MQTT heartbeat 消息更新
	BatteryVoltage          float64    `gorm:"type:decimal(6,2)" json:"battery_voltage"`                 // 主电池电压 (V)
	OnlineStatus            int8       `gorm:"type:tinyint;default:0" json:"online_status"`              // 在线状态: 0离线 1在线
	WorkStatus              int8       `gorm:"type:tinyint;default:0" json:"work_status"`                // 工作状态: 0空闲 1清扫中 2充电中 3故障 4维护
	FaultStatus             int8       `gorm:"type:tinyint;default:0" json:"fault_status"`               // 故障状态: 0正常 1异常
	WorkPeriod              int8       `gorm:"type:tinyint;default:0" json:"work_period"`                // 当前工作时段: 0无 1/2/3
	SignalStrength          int        `gorm:"type:int;default:0" json:"signal_strength"`                // 综合信号强度
	Signal4GStrength        int        `gorm:"type:int;default:0" json:"signal_4g_strength"`             // 4G 信号强度
	RunDurationSeconds      int64      `gorm:"type:bigint;default:0" json:"run_duration_seconds"`        // 累计运行时长 (秒)
	Speed                   float64    `gorm:"type:decimal(5,2)" json:"speed"`                           // 速度 (m/s)
	CleanArea               float64    `gorm:"type:decimal(10,2)" json:"clean_area"`                     // 累计清扫面积 (㎡)
	FaultCode               int        `gorm:"type:int;default:0" json:"fault_code"`                     // 故障码，0 表示无故障
	CartBatteryVoltage      float64    `gorm:"type:decimal(6,2)" json:"cart_battery_voltage"`            // 清扫小车电池电压 (V)
	CartLowVoltageStatus    int8       `gorm:"type:tinyint;default:0" json:"cart_low_voltage_status"`    // 清扫小车低电压状态: 0正常 1低压
	ShuttleBatteryVoltage   float64    `gorm:"type:decimal(6,2)" json:"shuttle_battery_voltage"`         // 接驳车/摆渡车电池电压 (V)
	ShuttleLowVoltageStatus int8       `gorm:"type:tinyint;default:0" json:"shuttle_low_voltage_status"` // 接驳车/摆渡车低电压状态
	ShuttleMotorCurrent     float64    `gorm:"type:decimal(6,2)" json:"shuttle_motor_current"`           // 接驳车电机电流 (A)
	CartCleanMotorCurrent   float64    `gorm:"type:decimal(6,2)" json:"cart_clean_motor_current"`        // 清扫小车清扫电机电流 (A)
	CartTravelMotorCurrent  float64    `gorm:"type:decimal(6,2)" json:"cart_travel_motor_current"`       // 清扫小车行进电机电流 (A)
	ShuttlePowerPercent     int8       `gorm:"type:tinyint;default:0" json:"shuttle_power_percent"`      // 接驳车功率选择 (%)
	ShuttleStartPosition    int8       `gorm:"type:tinyint;default:0" json:"shuttle_start_position"`     // 接驳车起点位置: 0否 1是
	ShuttleEndPosition      int8       `gorm:"type:tinyint;default:0" json:"shuttle_end_position"`       // 接驳车终点位置: 0否 1是
	ShuttleHasCleaner       int8       `gorm:"type:tinyint;default:0" json:"shuttle_has_cleaner"`        // 接驳车上是否有清扫车: 0否 1是
	Temperature             float64    `gorm:"type:decimal(5,2)" json:"temperature"`                     // 环境温度 (℃)
	Humidity                float64    `gorm:"type:decimal(5,2)" json:"humidity"`                        // 环境湿度 (%)
	LightIntensity          float64    `gorm:"type:decimal(10,2)" json:"light_intensity"`                // 光照强度 (lux)
	WindSpeed               float64    `gorm:"type:decimal(5,2)" json:"wind_speed"`                      // 风速 (m/s)
	DeviceTime              *time.Time `gorm:"type:datetime" json:"device_time"`                         // 设备当前时间
	LastHeartbeat           *time.Time `gorm:"type:datetime" json:"last_heartbeat"`                      // 最后心跳时间，用于判断在线状态
	CreateTime              time.Time  `gorm:"autoCreateTime" json:"create_time"`                        // 注册时间
	UpdateTime              time.Time  `gorm:"autoUpdateTime" json:"update_time"`                        // 最后更新时间
}

func (Robot) TableName() string {
	return "robots"
}
