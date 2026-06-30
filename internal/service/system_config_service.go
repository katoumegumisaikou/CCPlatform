package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// SystemConfigService 系统配置业务逻辑层。
type SystemConfigService struct {
	repo *repository.SystemConfigRepo
}

func NewSystemConfigService() *SystemConfigService {
	return &SystemConfigService{repo: repository.NewSystemConfigRepo()}
}

// GetByKey 根据键获取配置值。
func (s *SystemConfigService) GetByKey(key string) (*model.SystemConfig, error) {
	return s.repo.GetByKey(key)
}

// Set 设置配置项（不存在则创建，存在则更新）。
func (s *SystemConfigService) Set(cfg *model.SystemConfig) error {
	if cfg.ConfigKey == "" {
		return fmt.Errorf("config_key is required")
	}
	return s.repo.Set(cfg)
}

// List 分页查询配置列表。
func (s *SystemConfigService) List(page, size int, category string) ([]model.SystemConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, category)
}

// Delete 删除配置项。
func (s *SystemConfigService) Delete(key string) error {
	_, err := s.repo.GetByKey(key)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}
	return s.repo.Delete(key)
}

// GetAllByCategory 查询指定分类下的所有配置。
func (s *SystemConfigService) GetAllByCategory(category string) ([]model.SystemConfig, error) {
	return s.repo.GetAllByCategory(category)
}
