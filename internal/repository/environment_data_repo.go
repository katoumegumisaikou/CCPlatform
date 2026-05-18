package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// EnvironmentDataRepo 环境数据历史访问层。
type EnvironmentDataRepo struct {
	db *gorm.DB
}

func NewEnvironmentDataRepo() *EnvironmentDataRepo {
	return &EnvironmentDataRepo{db: DB}
}

// Create 写入一条环境数据记录。
func (r *EnvironmentDataRepo) Create(data *model.EnvironmentData) error {
	return r.db.Create(data).Error
}

// GetByRobotID 查询指定机器人环境数据历史。
func (r *EnvironmentDataRepo) GetByRobotID(robotID, startTime, endTime string) ([]model.EnvironmentData, error) {
	var data []model.EnvironmentData
	query := r.db.Model(&model.EnvironmentData{})
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			query = query.Where("record_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			query = query.Where("record_time <= ?", t)
		}
	}
	err := query.Order("record_time ASC").Limit(1000).Find(&data).Error
	return data, err
}
