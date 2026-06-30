package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// GeofenceRepo 电子围栏数据访问层，封装 geofences 表的所有数据库操作。
type GeofenceRepo struct {
	db *gorm.DB
}

// NewGeofenceRepo 创建 GeofenceRepo 实例，使用全局 DB 连接。
func NewGeofenceRepo() *GeofenceRepo {
	return &GeofenceRepo{db: DB}
}

// Create 创建电子围栏。
func (r *GeofenceRepo) Create(fence *model.Geofence) error {
	return r.db.Create(fence).Error
}

// GetByID 根据围栏 ID 查询。
func (r *GeofenceRepo) GetByID(id string) (*model.Geofence, error) {
	var f model.Geofence
	err := r.db.Where("geofence_id = ?", id).First(&f).Error
	return &f, err
}

// Update 更新电子围栏。
func (r *GeofenceRepo) Update(fence *model.Geofence) error {
	return r.db.Save(fence).Error
}

// Delete 删除电子围栏（物理删除）。
func (r *GeofenceRepo) Delete(id string) error {
	return r.db.Where("geofence_id = ?", id).Delete(&model.Geofence{}).Error
}

// List 分页查询电子围栏列表，支持按名称搜索和作用域筛选。
func (r *GeofenceRepo) List(page, size int, keyword string, scopeType string, fenceType int) ([]model.Geofence, int64, error) {
	var fences []model.Geofence
	var total int64
	query := r.db.Model(&model.Geofence{})
	if keyword != "" {
		query = query.Where("fence_name LIKE ?", "%"+keyword+"%")
	}
	if scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}
	if fenceType > 0 {
		query = query.Where("fence_type = ?", fenceType)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&fences).Error
	return fences, total, err
}

// GetAll 查询所有启用状态的电子围栏。
func (r *GeofenceRepo) GetAll() ([]model.Geofence, error) {
	var fences []model.Geofence
	err := r.db.Where("status = 1").Find(&fences).Error
	return fences, err
}

// GetByScope 查询指定作用域的电子围栏（如某个电站或机器人的所有围栏）。
func (r *GeofenceRepo) GetByScope(scopeType string, scopeID string) ([]model.Geofence, error) {
	var fences []model.Geofence
	err := r.db.Where("status = 1 AND scope_type = ? AND scope_id = ?", scopeType, scopeID).Find(&fences).Error
	return fences, err
}

// GetGlobalFences 查询全局作用域的电子围栏（scope_type=global）。
func (r *GeofenceRepo) GetGlobalFences() ([]model.Geofence, error) {
	var fences []model.Geofence
	err := r.db.Where("status = 1 AND scope_type = 'global'").Find(&fences).Error
	return fences, err
}
