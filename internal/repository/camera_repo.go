package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// CameraRepo 摄像头数据访问层。
type CameraRepo struct {
	db *gorm.DB
}

func NewCameraRepo() *CameraRepo {
	return &CameraRepo{db: DB}
}

func (r *CameraRepo) Create(cam *model.Camera) error {
	return r.db.Create(cam).Error
}

func (r *CameraRepo) GetByID(id string) (*model.Camera, error) {
	var cam model.Camera
	err := r.db.Where("camera_id = ?", id).First(&cam).Error
	return &cam, err
}

func (r *CameraRepo) Update(cam *model.Camera) error {
	return r.db.Save(cam).Error
}

func (r *CameraRepo) Delete(id string) error {
	return r.db.Where("camera_id = ?", id).Delete(&model.Camera{}).Error
}

func (r *CameraRepo) List(page, size int, stationID string) ([]model.Camera, int64, error) {
	var cams []model.Camera
	var total int64
	query := r.db.Model(&model.Camera{})
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&cams).Error
	return cams, total, err
}

func (r *CameraRepo) GetByStation(stationID string) ([]model.Camera, error) {
	var cams []model.Camera
	err := r.db.Where("station_id = ? AND status = 1", stationID).Find(&cams).Error
	return cams, err
}
