package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// DeviceCredentialRepo 设备认证凭据数据访问层。
type DeviceCredentialRepo struct {
	db *gorm.DB
}

func NewDeviceCredentialRepo() *DeviceCredentialRepo {
	return &DeviceCredentialRepo{db: DB}
}

func (r *DeviceCredentialRepo) Create(cred *model.DeviceCredential) error {
	return r.db.Create(cred).Error
}

func (r *DeviceCredentialRepo) GetByDeviceID(deviceID string) (*model.DeviceCredential, error) {
	var cred model.DeviceCredential
	err := r.db.Where("device_id = ?", deviceID).First(&cred).Error
	return &cred, err
}

func (r *DeviceCredentialRepo) GetByUsername(username string) (*model.DeviceCredential, error) {
	var cred model.DeviceCredential
	err := r.db.Where("username = ?", username).First(&cred).Error
	return &cred, err
}

func (r *DeviceCredentialRepo) Update(cred *model.DeviceCredential) error {
	return r.db.Save(cred).Error
}

func (r *DeviceCredentialRepo) Delete(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&model.DeviceCredential{}).Error
}
