package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// SystemConfigRepo 系统配置数据访问层，封装 system_configs 表的数据库操作。
type SystemConfigRepo struct {
	db *gorm.DB
}

func NewSystemConfigRepo() *SystemConfigRepo {
	return &SystemConfigRepo{db: DB}
}

// GetByKey 根据配置键查询单条配置。
func (r *SystemConfigRepo) GetByKey(key string) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&cfg).Error
	return &cfg, err
}

// Set 创建或更新配置项（config_key 存在则更新值，不存在则新增）。
func (r *SystemConfigRepo) Set(cfg *model.SystemConfig) error {
	var existing model.SystemConfig
	err := r.db.Where("config_key = ?", cfg.ConfigKey).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(cfg).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"config_value": cfg.ConfigValue,
		"category":     cfg.Category,
		"description":  cfg.Description,
	}).Error
}

// Delete 根据配置键删除配置。
func (r *SystemConfigRepo) Delete(key string) error {
	return r.db.Where("config_key = ?", key).Delete(&model.SystemConfig{}).Error
}

// List 分页查询配置列表，支持按分类筛选。
func (r *SystemConfigRepo) List(page, size int, category string) ([]model.SystemConfig, int64, error) {
	var configs []model.SystemConfig
	var total int64
	query := r.db.Model(&model.SystemConfig{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&configs).Error
	return configs, total, err
}

// GetAllByCategory 查询指定分类下的所有配置。
func (r *SystemConfigRepo) GetAllByCategory(category string) ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := r.db.Where("category = ?", category).Find(&configs).Error
	return configs, err
}
