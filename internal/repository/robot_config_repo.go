package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// RobotConfigRepo 设备配置模板数据访问层。
type RobotConfigRepo struct {
	db *gorm.DB
}

func NewRobotConfigRepo() *RobotConfigRepo {
	return &RobotConfigRepo{db: DB}
}

func (r *RobotConfigRepo) Create(cfg *model.RobotConfig) error {
	return r.db.Create(cfg).Error
}

func (r *RobotConfigRepo) GetByID(id string) (*model.RobotConfig, error) {
	var cfg model.RobotConfig
	err := r.db.Where("config_id = ?", id).First(&cfg).Error
	return &cfg, err
}

func (r *RobotConfigRepo) Update(cfg *model.RobotConfig) error {
	return r.db.Save(cfg).Error
}

func (r *RobotConfigRepo) Delete(id string) error {
	return r.db.Where("config_id = ?", id).Delete(&model.RobotConfig{}).Error
}

// List 分页查询配置模板，支持按机器人类型筛选。
func (r *RobotConfigRepo) List(page, size int, robotType int8) ([]model.RobotConfig, int64, error) {
	var configs []model.RobotConfig
	var total int64
	query := r.db.Model(&model.RobotConfig{})
	if robotType > 0 {
		query = query.Where("robot_type = ? OR robot_type = 0", robotType)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&configs).Error
	return configs, total, err
}

// GetDefault 获取指定机器人类型的默认配置模板。
func (r *RobotConfigRepo) GetDefault(robotType int8) (*model.RobotConfig, error) {
	var cfg model.RobotConfig
	err := r.db.Where("(robot_type = ? OR robot_type = 0) AND is_default = 1 AND status = 1", robotType).
		Order("robot_type DESC").First(&cfg).Error
	return &cfg, err
}
