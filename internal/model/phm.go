package model

import "time"

// PHM（故障预测与健康管理）覆盖的部件类型常量，用于 Component/PatternCode 等字段取值。
const (
	ComponentDriveMotor   = "drive_motor"  // 驱动电机
	ComponentBrushMotor   = "brush_motor"  // 滚刷电机
	ComponentBattery      = "battery"      // 电池组
	ComponentController   = "controller"   // 控制器
	ComponentSensor       = "sensor"       // 传感器
	ComponentTransmission = "transmission" // 传动系统
)

// 健康等级常量，对应百分制健康指数的四档颜色分级。
const (
	HealthLevelGreen  = "green"  // 80-100 健康
	HealthLevelYellow = "yellow" // 60-79 亚健康
	HealthLevelOrange = "orange" // 40-59 异常
	HealthLevelRed    = "red"    // 0-39 危险
)

// 故障风险等级常量。
const (
	RiskLevelHigh   = "high"   // 故障概率 > 70%
	RiskLevelMedium = "medium" // 40% - 70%
	RiskLevelLow    = "low"    // < 40%
)

// RobotComponentHealth 机器人部件健康指数当前状态，对应 robot_component_health 表。
// 每个机器人每个部件只保留一条最新记录（唯一联合索引），由 PHMService 周期计算后 upsert。
type RobotComponentHealth struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RobotID     string    `gorm:"type:varchar(32);uniqueIndex:idx_robot_component" json:"robot_id"`
	Component   string    `gorm:"type:varchar(50);uniqueIndex:idx_robot_component" json:"component"` // 部件类型，见 Component* 常量
	HealthScore float64   `gorm:"type:decimal(5,2)" json:"health_score"`                             // 健康指数 0-100
	HealthLevel string    `gorm:"type:varchar(20)" json:"health_level"`                              // 健康等级，见 HealthLevel* 常量
	DegradeRate float64   `gorm:"type:decimal(6,3)" json:"degrade_rate"`                             // 健康度日均下降速率（分/天），用于故障窗口预测
	Metrics     string    `gorm:"type:text" json:"metrics"`                                          // 细分指标快照（JSON），用于前端展示评分依据
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (RobotComponentHealth) TableName() string {
	return "robot_component_health"
}

// RobotHealthHistory 机器人部件健康指数历史快照，对应 robot_health_history 表。
// 每次健康度计算追加一条记录，用于趋势曲线、同比/环比分析和退化速率回归。
type RobotHealthHistory struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RobotID     string    `gorm:"type:varchar(32);index:idx_health_hist_robot_comp" json:"robot_id"`
	Component   string    `gorm:"type:varchar(50);index:idx_health_hist_robot_comp" json:"component"`
	HealthScore float64   `gorm:"type:decimal(5,2)" json:"health_score"`
	RecordTime  time.Time `gorm:"type:datetime;index" json:"record_time"`
}

func (RobotHealthHistory) TableName() string {
	return "robot_health_history"
}

// RobotSensorData 机器人 PHM 专用传感器数据（振动/温度/电气参数），对应 robot_sensor_data 表。
// 由 MQTT sensor 消息写入，是振动频谱分析、绝缘电阻监测等指标的原始/派生数据来源。
type RobotSensorData struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RobotID              string    `gorm:"type:varchar(32);index" json:"robot_id"`
	Component            string    `gorm:"type:varchar(50);index" json:"component"`          // 采集部件，见 Component* 常量
	VibrationRMS         float64   `gorm:"type:decimal(8,4)" json:"vibration_rms"`           // 振动有效值 (mm/s)
	VibrationPeak        float64   `gorm:"type:decimal(8,4)" json:"vibration_peak"`          // 振动峰值 (mm/s)
	CrestFactor          float64   `gorm:"type:decimal(6,3)" json:"crest_factor"`            // 峰值因子 = 峰值/RMS，轴承早期故障敏感指标
	DominantFreqHz       float64   `gorm:"type:decimal(8,2)" json:"dominant_freq_hz"`        // 振动信号主频 (Hz)
	BearingFaultFlag     string    `gorm:"type:varchar(20)" json:"bearing_fault_flag"`       // 轴承故障特征: none/outer_race/inner_race/ball
	WindingTemp          float64   `gorm:"type:decimal(5,2)" json:"winding_temp"`            // 电机绕组温度 (℃)
	ControllerTemp       float64   `gorm:"type:decimal(5,2)" json:"controller_temp"`         // 控制器散热片温度 (℃)
	InsulationResistMOhm float64   `gorm:"type:decimal(10,2)" json:"insulation_resist_mohm"` // 绝缘电阻 (MΩ)，越低老化/漏电风险越高
	RecordTime           time.Time `gorm:"type:datetime;index" json:"record_time"`
}

func (RobotSensorData) TableName() string {
	return "robot_sensor_data"
}

// FaultPattern 故障模式库（静态知识库），对应 fault_patterns 表。
// 覆盖电机异常、电池衰减、传感器漂移、通信中断等故障类别，由 PHMService 启动时做幂等种子初始化。
type FaultPattern struct {
	PatternCode      string `gorm:"primaryKey;type:varchar(50)" json:"pattern_code"`
	Component        string `gorm:"type:varchar(50);index" json:"component"`    // 所属部件，见 Component* 常量
	FaultName        string `gorm:"type:varchar(100)" json:"fault_name"`        // 故障名称
	Category         string `gorm:"type:varchar(50)" json:"category"`           // 故障类别: motor/battery/sensor/comm/transmission
	SymptomDesc      string `gorm:"type:varchar(500)" json:"symptom_desc"`      // 典型症状描述
	SuggestedAction  string `gorm:"type:varchar(500)" json:"suggested_action"`  // 建议处置方案: 维修/更换/校准
	RecommendedParts string `gorm:"type:varchar(200)" json:"recommended_parts"` // 推荐备件（型号，逗号分隔）
}

func (FaultPattern) TableName() string {
	return "fault_patterns"
}

// FaultPrediction 故障预测结果记录，对应 fault_predictions 表。
// 由 PHMService 基于健康指数退化趋势匹配故障模式库生成，是维护工单和备件预测的数据源。
type FaultPrediction struct {
	PredictionID     string    `gorm:"primaryKey;type:varchar(32)" json:"prediction_id"`
	RobotID          string    `gorm:"type:varchar(32);index" json:"robot_id"`
	StationID        string    `gorm:"type:varchar(32);index" json:"station_id"`
	Component        string    `gorm:"type:varchar(50)" json:"component"`
	PatternCode      string    `gorm:"type:varchar(50)" json:"pattern_code"`
	FaultName        string    `gorm:"type:varchar(100)" json:"fault_name"`
	Probability      float64   `gorm:"type:decimal(5,4)" json:"probability"` // 故障概率 0-1
	RiskLevel        string    `gorm:"type:varchar(20)" json:"risk_level"`   // 见 RiskLevel* 常量
	Confidence       float64   `gorm:"type:decimal(5,4)" json:"confidence"`  // 预测置信度 0-1
	PredictedStart   time.Time `gorm:"type:datetime" json:"predicted_start"` // 预测故障窗口开始时间
	PredictedEnd     time.Time `gorm:"type:datetime" json:"predicted_end"`   // 预测故障窗口结束时间
	SuggestedAction  string    `gorm:"type:varchar(500)" json:"suggested_action"`
	RecommendedParts string    `gorm:"type:varchar(200)" json:"recommended_parts"`
	Status           int8      `gorm:"type:tinyint;default:0;index" json:"status"` // 0待处理 1已生成工单 2已忽略
	MaintID          string    `gorm:"type:varchar(32)" json:"maint_id"`           // 关联生成的维护工单 ID
	CreateTime       time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime       time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (FaultPrediction) TableName() string {
	return "fault_predictions"
}
