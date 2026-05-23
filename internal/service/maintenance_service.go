package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"fmt"
)

// MaintenanceService 维护记录业务逻辑层。
type MaintenanceService struct {
	repo *repository.MaintenanceRepo
}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{repo: repository.NewMaintenanceRepo()}
}

func (s *MaintenanceService) Create(m *model.Maintenance) error {
	m.MaintID = util.GenerateID()
	return s.repo.Create(m)
}

func (s *MaintenanceService) Update(m *model.Maintenance) error {
	_, err := s.repo.GetByID(m.MaintID)
	if err != nil {
		return fmt.Errorf("maintenance record not found: %w", err)
	}
	return s.repo.Update(m)
}

func (s *MaintenanceService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *MaintenanceService) List(page, size int, robotID, maintType, startTime, endTime string) ([]model.Maintenance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, robotID, maintType, startTime, endTime)
}

// GetReminders 查询未来 7 天内的维护提醒。
func (s *MaintenanceService) GetReminders() ([]model.Maintenance, error) {
	return s.repo.GetUpcomingReminders(7)
}

// GetByRobotID 查询指定机器人的维护历史。
func (s *MaintenanceService) GetByRobotID(robotID string, page, size int) ([]model.Maintenance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.GetByRobotID(robotID, page, size)
}
