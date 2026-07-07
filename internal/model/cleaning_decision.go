package model

import "time"

// CleaningDecisionRule 清扫决策规则，定义气象条件下是否建议清扫的阈值。
// 支持按电站设置专属规则（station_id 非空）和全局规则（station_id 为空）。
type CleaningDecisionRule struct {
	RuleID     string    `gorm:"primaryKey;type:varchar(32)" json:"rule_id"`          // 规则唯一标识
	RuleName   string    `gorm:"type:varchar(100);not null" json:"rule_name"`          // 规则名称
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`             // 所属电站（空 = 全局规则）
	Conditions string    `gorm:"type:text" json:"conditions"`                          // JSON: 气象条件阈值 {"max_wind_scale":5,"min_temp":-10,"max_temp":45,"max_humidity":90,"no_rain":true,"min_vis_km":5}
	DustRule   string    `gorm:"type:text" json:"dust_rule"`                           // JSON: 灰尘累积规则 {"days_since_last_clean":3,"dust_threshold":500}
	Priority   int8      `gorm:"type:tinyint;default:1" json:"priority"`               // 优先级 (1-5, 5 最高)
	Status     int8      `gorm:"type:tinyint;default:1" json:"status"`                 // 状态: 0停用 1启用
	Remark     string    `gorm:"type:varchar(500)" json:"remark"`                      // 备注说明
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (CleaningDecisionRule) TableName() string { return "cleaning_decision_rules" }

// ExtremeWeatherRule 极端天气应对规则，定义触发条件与响应动作。
type ExtremeWeatherRule struct {
	RuleID       string    `gorm:"primaryKey;type:varchar(32)" json:"rule_id"`
	RuleName     string    `gorm:"type:varchar(100);not null" json:"rule_name"`
	StationID    string    `gorm:"type:varchar(32);index" json:"station_id"`        // 所属电站（空 = 全局规则）
	WeatherType  string    `gorm:"type:varchar(50)" json:"weather_type"`            // 天气类型: 暴雨/大风/暴雪/高温/沙尘暴
	TriggerCond  string    `gorm:"type:text" json:"trigger_cond"`                   // JSON: 触发条件 {"wind_scale_gt":6,"precip_gt":50,"warning_level":"red"}
	Action       string    `gorm:"type:varchar(50)" json:"action"`                  // 响应动作: 暂停清扫/返回充电/降低速度
	RecoveryCond string    `gorm:"type:text" json:"recovery_cond"`                  // JSON: 恢复条件 {"wind_scale_lt":4,"precip_lt":1}
	Status       int8      `gorm:"type:tinyint;default:1" json:"status"`
	Remark       string    `gorm:"type:varchar(500)" json:"remark"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ExtremeWeatherRule) TableName() string { return "extreme_weather_rules" }

// CleaningScheduleAdjustment 动态频次调整建议记录，记录系统分析结果和用户采纳情况。
type CleaningScheduleAdjustment struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StationID     string    `gorm:"type:varchar(32);index" json:"station_id"` // 电站 ID
	CurrentFreq   string    `gorm:"type:varchar(50)" json:"current_freq"`     // 当前清扫频次（如 "每3天一次"）
	SuggestedFreq string    `gorm:"type:varchar(50)" json:"suggested_freq"`   // 建议清扫频次
	Reason        string    `gorm:"type:varchar(500)" json:"reason"`          // 调整原因
	AnalysisData  string    `gorm:"type:text" json:"analysis_data"`           // JSON: 分析数据 {"dust_accumulation_rate":150,"weather_score":0.85,"next_7d_clean_windows":5}
	Status        int8      `gorm:"type:tinyint;default:1" json:"status"`     // 状态: 1建议 2已采纳 3已忽略
	CreateTime    time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CleaningScheduleAdjustment) TableName() string { return "cleaning_schedule_adjustments" }
