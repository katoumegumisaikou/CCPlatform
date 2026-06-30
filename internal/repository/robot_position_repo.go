package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// RobotPositionRepo 机器人位置历史数据访问层。
type RobotPositionRepo struct {
	db *gorm.DB
}

func NewRobotPositionRepo() *RobotPositionRepo {
	return &RobotPositionRepo{db: DB}
}

// Create 写入一条位置历史记录。
func (r *RobotPositionRepo) Create(pos *model.RobotPosition) error {
	return r.db.Create(pos).Error
}

// GetByRobotID 查询指定机器人位置历史，用于轨迹回放。
func (r *RobotPositionRepo) GetByRobotID(robotID, startTime, endTime string, limit int) ([]model.RobotPosition, error) {
	var positions []model.RobotPosition
	query := r.db.Model(&model.RobotPosition{})
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			query = query.Where("timestamp >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			query = query.Where("timestamp <= ?", t)
		}
	}
	if limit <= 0 {
		limit = 1000
	}
	err := query.Order("timestamp ASC").Limit(limit).Find(&positions).Error
	return positions, err
}

