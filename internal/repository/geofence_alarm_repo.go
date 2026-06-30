package repository

import (
	"ccplatform/internal/model"
	"time"

	"gorm.io/gorm"
)

// GeofenceAlarmRepo 围栏告警数据访问层，封装 geofence_alarms 表的操作。
type GeofenceAlarmRepo struct {
	db *gorm.DB
}

// NewGeofenceAlarmRepo 创建 GeofenceAlarmRepo 实例。
func NewGeofenceAlarmRepo() *GeofenceAlarmRepo {
	return &GeofenceAlarmRepo{db: DB}
}

// Create 创建围栏告警记录。
func (r *GeofenceAlarmRepo) Create(alarm *model.GeofenceAlarm) error {
	return r.db.Create(alarm).Error
}

// GetByID 根据告警 ID 查询。
func (r *GeofenceAlarmRepo) GetByID(id string) (*model.GeofenceAlarm, error) {
	var a model.GeofenceAlarm
	err := r.db.Where("alarm_id = ?", id).First(&a).Error
	return &a, err
}

// Update 更新告警处理状态。
func (r *GeofenceAlarmRepo) Update(alarm *model.GeofenceAlarm) error {
	return r.db.Save(alarm).Error
}

// List 分页查询围栏告警记录，支持多维度筛选。
func (r *GeofenceAlarmRepo) List(page, size int, fenceID, robotID, stationID string, triggerType, handleStatus int) ([]model.GeofenceAlarm, int64, error) {
	var alarms []model.GeofenceAlarm
	var total int64
	query := r.db.Model(&model.GeofenceAlarm{})
	if fenceID != "" {
		query = query.Where("fence_id = ?", fenceID)
	}
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if triggerType > 0 {
		query = query.Where("trigger_type = ?", triggerType)
	}
	if handleStatus >= 0 {
		query = query.Where("handle_status = ?", handleStatus)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("alarm_time DESC").Find(&alarms).Error
	return alarms, total, err
}

// GetRecent 查询最近 N 条未处理的围栏告警。
func (r *GeofenceAlarmRepo) GetRecent(limit int) ([]model.GeofenceAlarm, error) {
	var alarms []model.GeofenceAlarm
	err := r.db.Where("handle_status = 0").
		Order("alarm_time DESC").
		Limit(limit).
		Find(&alarms).Error
	return alarms, err
}

// HasRecentAlarm 检查指定机器人在最近 cooldownSec 秒内是否已有同类型告警（防重复）。
func (r *GeofenceAlarmRepo) HasRecentAlarm(fenceID, robotID string, triggerType int8, cooldownSec int) (bool, error) {
	var count int64
	cutoff := time.Now().Add(-time.Duration(cooldownSec) * time.Second)
	err := r.db.Model(&model.GeofenceAlarm{}).
		Where("fence_id = ? AND robot_id = ? AND trigger_type = ? AND alarm_time >= ?",
			fenceID, robotID, triggerType, cutoff).
		Count(&count).Error
	return count > 0, err
}
