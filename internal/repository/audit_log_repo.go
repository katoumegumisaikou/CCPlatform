package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// AuditLogRepo 操作日志数据访问层。
type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo() *AuditLogRepo {
	return &AuditLogRepo{db: DB}
}

func (r *AuditLogRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// List 分页查询操作日志，支持按操作人、操作类型、时间段筛选。
func (r *AuditLogRepo) List(page, size int, operatorID, operationType, startTime, endTime string) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})
	if operatorID != "" {
		query = query.Where("operator_id = ?", operatorID)
	}
	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}
	if startTime != "" {
		query = query.Where("create_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("create_time <= ?", endTime)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&logs).Error
	return logs, total, err
}
