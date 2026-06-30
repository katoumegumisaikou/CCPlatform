package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// StationRepo 电站数据访问层，封装 stations 表的所有数据库操作。
type StationRepo struct {
	db *gorm.DB
}

// NewStationRepo 创建 StationRepo 实例，使用全局 DB 连接。
func NewStationRepo() *StationRepo {
	return &StationRepo{db: DB}
}

// Create 创建新电站记录。
func (r *StationRepo) Create(s *model.Station) error {
	return r.db.Create(s).Error
}

// GetByID 根据电站 ID 查询单条记录。
func (r *StationRepo) GetByID(id string) (*model.Station, error) {
	var s model.Station
	err := r.db.Where("station_id = ?", id).First(&s).Error
	return &s, err
}

// Update 更新电站信息（全量更新）。
func (r *StationRepo) Update(s *model.Station) error {
	return r.db.Save(s).Error
}

// Delete 根据电站 ID 删除记录（物理删除）。
func (r *StationRepo) Delete(id string) error {
	return r.db.Where("station_id = ?", id).Delete(&model.Station{}).Error
}

// List 分页查询电站列表，支持按名称或编码模糊搜索。
// 返回电站列表、总数和错误信息。keyword 为空时不进行过滤。
func (r *StationRepo) List(page, size int, keyword string) ([]model.Station, int64, error) {
	var stations []model.Station
	var total int64
	query := r.db.Model(&model.Station{})
	if keyword != "" {
		query = query.Where("station_name LIKE ? OR station_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&stations).Error
	return stations, total, err
}

// GetAll 查询所有启用状态的电站（status=1），用于下拉选择等场景。
func (r *StationRepo) GetAll() ([]model.Station, error) {
	var stations []model.Station
	err := r.db.Where("status = 1").Find(&stations).Error
	return stations, err
}
