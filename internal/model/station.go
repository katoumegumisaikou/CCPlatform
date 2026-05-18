// Package model 定义平台所有数据模型，对应 MySQL 数据库表结构。
// 使用 GORM 标签控制数据库列属性，使用 JSON 标签控制 API 序列化。
package model

import "time"

// Station 光伏电站模型，对应 stations 表。
// 一个电站包含多台机器人，是平台多租户架构的基本单位。
type Station struct {
	StationID   string    `gorm:"primaryKey;type:varchar(32)" json:"station_id"`             // 电站唯一标识
	StationName string    `gorm:"type:varchar(100);not null" json:"station_name"`            // 电站名称
	StationCode string    `gorm:"type:varchar(50);uniqueIndex" json:"station_code"`          // 电站编码（唯一）
	Location    string    `gorm:"type:varchar(200)" json:"location"`                         // 电站地址
	Longitude   float64   `gorm:"type:decimal(10,7)" json:"longitude"`                       // 经度
	Latitude    float64   `gorm:"type:decimal(10,7)" json:"latitude"`                        // 纬度
	Capacity    float64   `gorm:"type:decimal(10,2)" json:"capacity"`                        // 装机容量 (MW)
	PanelArea   float64   `gorm:"type:decimal(12,2)" json:"panel_area"`                      // 光伏板总面积 (㎡)
	PanelCount  int       `gorm:"type:int" json:"panel_count"`                               // 光伏板数量
	Status      int8      `gorm:"type:tinyint;default:1" json:"status"`                      // 状态: 0停用 1启用
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`                         // 创建时间
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`                         // 更新时间
}

func (Station) TableName() string {
	return "stations"
}
