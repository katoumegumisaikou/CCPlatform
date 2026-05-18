// Package service 提供业务逻辑层，处理数据校验、业务规则和跨模块协调。
// 每个 Service 结构体封装对应模块的业务逻辑，调用 Repository 层进行数据操作。
package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"crypto/rand"
	"fmt"
)

// StationService 电站业务逻辑层，处理电站的增删改查操作。
//
// TODO: 跨站对比分析 — 多电站数据横向对比 (需求 4.5.2)
// TODO: 电站组织归属 — 与组织架构模型关联 (需求 4.6.2)
type StationService struct {
	repo *repository.StationRepo
}

// NewStationService 创建 StationService 实例。
func NewStationService() *StationService {
	return &StationService{repo: repository.NewStationRepo()}
}

// Create 创建新电站，自动生成 8 位十六进制 ID。
func (s *StationService) Create(station *model.Station) error {
	station.StationID = generateID()
	return s.repo.Create(station)
}

// generateID 生成 8 位随机十六进制字符串作为业务 ID。
// 使用 crypto/rand 保证随机性，适用于所有实体的 ID 生成。
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GetByID 根据 ID 查询电站详情。
func (s *StationService) GetByID(id string) (*model.Station, error) {
	return s.repo.GetByID(id)
}

// Update 更新电站信息。
func (s *StationService) Update(station *model.Station) error {
	return s.repo.Update(station)
}

// Delete 删除电站。
func (s *StationService) Delete(id string) error {
	return s.repo.Delete(id)
}

// List 分页查询电站列表，支持关键词搜索。page/size 无效时使用默认值 1/10。
func (s *StationService) List(page, size int, keyword string) ([]model.Station, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, keyword)
}

// GetAll 查询所有启用状态的电站，用于下拉选择框。
func (s *StationService) GetAll() ([]model.Station, error) {
	return s.repo.GetAll()
}
