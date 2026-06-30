package model

import "time"

// AuditLog 操作日志模型，对应 audit_logs 表。
// 记录用户的关键操作，用于安全审计和追溯。
type AuditLog struct {
	LogID          string    `gorm:"primaryKey;type:varchar(32)" json:"log_id"`
	OperatorID     string    `gorm:"type:varchar(32);index" json:"operator_id"`
	OperatorName   string    `gorm:"type:varchar(50)" json:"operator_name"`
	OperationType  string    `gorm:"type:varchar(50);index" json:"operation_type"` // 操作类型: create/update/delete/login
	TargetResource string    `gorm:"type:varchar(200)" json:"target_resource"`     // 目标资源路径
	Detail         string    `gorm:"type:varchar(1000)" json:"detail"`              // 操作详情
	IP             string    `gorm:"type:varchar(50)" json:"ip"`                    // 操作IP
	CreateTime     time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
