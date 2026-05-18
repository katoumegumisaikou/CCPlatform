// Package model 定义平台所有数据模型，对应 MySQL 数据库表结构。
// 使用 GORM 标签控制数据库列属性，使用 JSON 标签控制 API 序列化。
//
// TODO: 固件模型 (Firmware) — 固件版本、升级包路径、兼容机器人类型、升级日志 (需求 4.3)
// TODO: 维护记录模型 (Maintenance) — 维护时间、维护类型、处理人、故障描述、配件更换 (需求 4.3)
// TODO: 操作日志模型 (AuditLog) — 操作人、操作类型、目标资源、操作详情、IP、时间 (需求 4.6.2)
// TODO: 登录日志模型 (LoginLog) — 用户、登录IP、登录时间、登录结果 (需求 4.6.2)
// TODO: 系统配置模型 (SystemConfig) — 配置键值对、分类、描述 (需求 4.6.2)
// TODO: 数据字典模型 (Dict) — 字典类型、字典项、排序 (需求 4.6.2)
// TODO: 组织架构模型 (Organization) — 多级组织结构，公司/区域/电站层级 (需求 4.6.2)
// TODO: 告警规则模型 (AlarmRule) — 阈值配置、升级规则、抑制规则 (需求 4.4.2)
// TODO: 通知模板模型 (NotifyTemplate) — 短信/邮件/APP 推送模板 (需求 4.4.2)
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
