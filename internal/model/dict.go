package model

import "time"

// Dict 数据字典类型模型，对应 dicts 表。
// 定义字典分类，如 alarm_type、robot_type 等。
type Dict struct {
	DictID      string    `gorm:"primaryKey;type:varchar(32)" json:"dict_id"`
	DictType    string    `gorm:"type:varchar(50);uniqueIndex" json:"dict_type"` // 字典类型编码
	DictName    string    `gorm:"type:varchar(100);not null" json:"dict_name"`   // 字典类型名称
	Description string    `gorm:"type:varchar(200)" json:"description"`
	Status      int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (Dict) TableName() string {
	return "dicts"
}

// DictItem 数据字典项模型，对应 dict_items 表。
// 每个字典类型下包含多个字典项，如 alarm_type 下的 MOTOR_FAULT、SENSOR_ERROR 等。
type DictItem struct {
	ItemID    string    `gorm:"primaryKey;type:varchar(32)" json:"item_id"`
	DictType  string    `gorm:"type:varchar(50);index" json:"dict_type"`       // 所属字典类型
	ItemKey   string    `gorm:"type:varchar(50)" json:"item_key"`              // 字典项键
	ItemValue string    `gorm:"type:varchar(200)" json:"item_value"`           // 字典项值
	SortOrder int       `gorm:"type:int;default:0" json:"sort_order"`          // 排序
	Status    int8      `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (DictItem) TableName() string {
	return "dict_items"
}
