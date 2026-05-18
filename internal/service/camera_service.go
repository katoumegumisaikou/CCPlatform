package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
)

// CameraService 摄像头管理业务逻辑层。
type CameraService struct {
	repo *repository.CameraRepo
}

func NewCameraService() *CameraService {
	return &CameraService{repo: repository.NewCameraRepo()}
}

func (s *CameraService) Create(cam *model.Camera) error {
	cam.CameraID = generateID()
	return s.repo.Create(cam)
}

func (s *CameraService) Update(cam *model.Camera) error {
	return s.repo.Update(cam)
}

func (s *CameraService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *CameraService) List(page, size int, stationID string) ([]model.Camera, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, stationID)
}

func (s *CameraService) GetByStation(stationID string) ([]model.Camera, error) {
	return s.repo.GetByStation(stationID)
}
