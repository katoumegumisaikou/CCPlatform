package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// FirmwareRepo 固件数据访问层。
type FirmwareRepo struct {
	db *gorm.DB
}

func NewFirmwareRepo() *FirmwareRepo {
	return &FirmwareRepo{db: DB}
}

func (r *FirmwareRepo) Create(fw *model.Firmware) error {
	return r.db.Create(fw).Error
}

func (r *FirmwareRepo) GetByID(id string) (*model.Firmware, error) {
	var fw model.Firmware
	err := r.db.Where("firmware_id = ?", id).First(&fw).Error
	return &fw, err
}

func (r *FirmwareRepo) Update(fw *model.Firmware) error {
	return r.db.Save(fw).Error
}

func (r *FirmwareRepo) Delete(id string) error {
	return r.db.Where("firmware_id = ?", id).Delete(&model.Firmware{}).Error
}

// List 分页查询固件列表，支持按兼容类型筛选。
func (r *FirmwareRepo) List(page, size int, compatibleType string) ([]model.Firmware, int64, error) {
	var firmwares []model.Firmware
	var total int64
	query := r.db.Model(&model.Firmware{})
	if compatibleType != "" {
		query = query.Where("compatible_robot_types LIKE ?", "%"+compatibleType+"%")
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&firmwares).Error
	return firmwares, total, err
}

// GetLatest 获取指定机器人类型的最新固件版本。
func (r *FirmwareRepo) GetLatest(robotType string) (*model.Firmware, error) {
	var fw model.Firmware
	err := r.db.Where("compatible_robot_types LIKE ? AND status = 1", "%"+robotType+"%").
		Order("create_time DESC").First(&fw).Error
	return &fw, err
}
