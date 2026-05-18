package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// LoginLogRepo 登录日志数据访问层。
type LoginLogRepo struct {
	db *gorm.DB
}

func NewLoginLogRepo() *LoginLogRepo {
	return &LoginLogRepo{db: DB}
}

func (r *LoginLogRepo) Create(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

// List 分页查询登录日志，支持按用户、时间段筛选。
func (r *LoginLogRepo) List(page, size int, userID, startTime, endTime string) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64
	query := r.db.Model(&model.LoginLog{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if startTime != "" {
		query = query.Where("login_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("login_time <= ?", endTime)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("login_time DESC").Find(&logs).Error
	return logs, total, err
}
