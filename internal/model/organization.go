package model

import "time"

// Organization 组织架构模型，对应 organizations 表。
// 支持多级组织结构（公司/区域/电站层级），通过 ParentID 构建树形关系。
type Organization struct {
	OrgID      string    `gorm:"primaryKey;type:varchar(32)" json:"org_id"`
	OrgName    string    `gorm:"type:varchar(100);not null" json:"org_name"`
	OrgCode    string    `gorm:"type:varchar(50);uniqueIndex" json:"org_code"`
	OrgType    string    `gorm:"type:varchar(20)" json:"org_type"`       // 组织类型: company/region/station_group
	ParentID   string    `gorm:"type:varchar(32);index" json:"parent_id"` // 父组织ID，空表示根节点
	StationIDs string    `gorm:"type:varchar(500)" json:"station_ids"`   // 关联电站ID列表 (逗号分隔)
	SortOrder  int       `gorm:"type:int;default:0" json:"sort_order"`
	Status     int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (Organization) TableName() string {
	return "organizations"
}
