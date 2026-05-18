package model

import "time"

// CleaningRecord 清扫记录模型，对应 cleaning_records 表。
// 每次 MQTT clean 消息上报时写入一条记录，用于真实效率计算。
type CleaningRecord struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID     uint      `gorm:"index" json:"task_id"`
	RobotID    string    `gorm:"type:varchar(32);index" json:"robot_id"`
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`
	PosX       float64   `gorm:"type:decimal(10,3)" json:"pos_x"`
	PosY       float64   `gorm:"type:decimal(10,3)" json:"pos_y"`
	CleanArea  float64   `gorm:"type:decimal(10,2)" json:"clean_area"`
	RecordTime time.Time `gorm:"type:datetime;index" json:"record_time"`
}

func (CleaningRecord) TableName() string {
	return "cleaning_records"
}

// ReportTemplate 自定义报表模板模型。
type ReportTemplate struct {
	TplID      string    `gorm:"primaryKey;type:varchar(32)" json:"tpl_id"`
	TplName    string    `gorm:"type:varchar(100);not null" json:"tpl_name"`
	TplType    string    `gorm:"type:varchar(20)" json:"tpl_type"`   // efficiency/economic/custom
	Dimensions string    `gorm:"type:text" json:"dimensions"`        // JSON: ["station","robot","date"]
	Metrics    string    `gorm:"type:text" json:"metrics"`           // JSON: ["clean_area","coverage_rate"]
	Status     int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (ReportTemplate) TableName() string {
	return "report_templates"
}
