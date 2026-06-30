package model

import "time"

// EnvironmentData 环境数据历史记录模型，对应 environment_data 表。
// 每次 MQTT status 消息中的环境参数同步写入，用于趋势分析。
type EnvironmentData struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RobotID        string    `gorm:"type:varchar(32);index" json:"robot_id"`
	Temperature    float64   `gorm:"type:decimal(5,2)" json:"temperature"`
	Humidity       float64   `gorm:"type:decimal(5,2)" json:"humidity"`
	LightIntensity float64   `gorm:"type:decimal(10,2)" json:"light_intensity"`
	WindSpeed      float64   `gorm:"type:decimal(5,2)" json:"wind_speed"`
	RecordTime     time.Time `gorm:"type:datetime;index" json:"record_time"`
}

func (EnvironmentData) TableName() string {
	return "environment_data"
}
