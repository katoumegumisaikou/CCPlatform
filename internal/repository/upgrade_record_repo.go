package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// UpgradeRecordRepo OTA 升级记录数据访问层。
type UpgradeRecordRepo struct {
	db *gorm.DB
}

func NewUpgradeRecordRepo() *UpgradeRecordRepo {
	return &UpgradeRecordRepo{db: DB}
}

func (r *UpgradeRecordRepo) Create(rec *model.UpgradeRecord) error {
	return r.db.Create(rec).Error
}

func (r *UpgradeRecordRepo) Update(rec *model.UpgradeRecord) error {
	return r.db.Save(rec).Error
}

func (r *UpgradeRecordRepo) GetByRobotID(robotID string) ([]model.UpgradeRecord, error) {
	var records []model.UpgradeRecord
	err := r.db.Where("robot_id = ?", robotID).Order("create_time DESC").Find(&records).Error
	return records, err
}

func (r *UpgradeRecordRepo) GetByFirmwareID(firmwareID string) ([]model.UpgradeRecord, error) {
	var records []model.UpgradeRecord
	err := r.db.Where("firmware_id = ?", firmwareID).Order("create_time DESC").Find(&records).Error
	return records, err
}
