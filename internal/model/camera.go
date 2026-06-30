package model

import "time"

// Camera 摄像头模型，对应 cameras 表。
type Camera struct {
	CameraID   string    `gorm:"primaryKey;type:varchar(32)" json:"camera_id"`
	CameraName string    `gorm:"type:varchar(100);not null" json:"camera_name"`
	StationID  string    `gorm:"type:varchar(32);index" json:"station_id"`
	StreamURL  string    `gorm:"type:varchar(500)" json:"stream_url"`
	CameraType string    `gorm:"type:varchar(20)" json:"camera_type"` // rtsp/hls
	Position   string    `gorm:"type:varchar(100)" json:"position"`   // 安装位置描述
	Status     int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (Camera) TableName() string {
	return "cameras"
}
