package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// MaintenanceRepo 维护记录数据访问层。
type MaintenanceRepo struct {
	db *gorm.DB
}

func NewMaintenanceRepo() *MaintenanceRepo {
	return &MaintenanceRepo{db: DB}
}

func (r *MaintenanceRepo) Create(m *model.Maintenance) error {
	return r.db.Create(m).Error
}

func (r *MaintenanceRepo) GetByID(id string) (*model.Maintenance, error) {
	var m model.Maintenance
	err := r.db.Where("maint_id = ?", id).First(&m).Error
	return &m, err
}

func (r *MaintenanceRepo) Update(m *model.Maintenance) error {
	return r.db.Save(m).Error
}

func (r *MaintenanceRepo) Delete(id string) error {
	return r.db.Where("maint_id = ?", id).Delete(&model.Maintenance{}).Error
}

// List 分页查询维护记录，支持按机器人、类型、时间段筛选。
func (r *MaintenanceRepo) List(page, size int, robotID, maintType, startTime, endTime string) ([]model.Maintenance, int64, error) {
	var records []model.Maintenance
	var total int64
	query := r.db.Model(&model.Maintenance{})
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if maintType != "" {
		query = query.Where("maint_type = ?", maintType)
	}
	if startTime != "" {
		query = query.Where("maint_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("maint_time <= ?", endTime)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("maint_time DESC").Find(&records).Error
	return records, total, err
}

// GetUpcomingReminders 查询未来 N 天内需要维护的提醒。
func (r *MaintenanceRepo) GetUpcomingReminders(days int) ([]model.Maintenance, error) {
	var records []model.Maintenance
	now := time.Now()
	deadline := now.AddDate(0, 0, days)
	err := r.db.Where("next_maint_time BETWEEN ? AND ?", now, deadline).Order("next_maint_time ASC").Find(&records).Error
	return records, err
}

// GetByRobotID 查询指定机器人的所有维护记录。
func (r *MaintenanceRepo) GetByRobotID(robotID string, page, size int) ([]model.Maintenance, int64, error) {
	var records []model.Maintenance
	var total int64
	r.db.Model(&model.Maintenance{}).Where("robot_id = ?", robotID).Count(&total)
	err := r.db.Where("robot_id = ?", robotID).
		Offset((page - 1) * size).Limit(size).Order("maint_time DESC").Find(&records).Error
	return records, total, err
}
