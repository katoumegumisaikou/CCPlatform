package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"fmt"
)

// FirmwareService 固件管理业务逻辑层。
type FirmwareService struct {
	repo      *repository.FirmwareRepo
	robotRepo *repository.RobotRepo
	publisher *mqtt.Publisher
}

func NewFirmwareService(publisher *mqtt.Publisher) *FirmwareService {
	return &FirmwareService{
		repo:      repository.NewFirmwareRepo(),
		robotRepo: repository.NewRobotRepo(),
		publisher: publisher,
	}
}

func (s *FirmwareService) Create(fw *model.Firmware) error {
	fw.FirmwareID = util.GenerateID()
	return s.repo.Create(fw)
}

func (s *FirmwareService) Update(fw *model.Firmware) error {
	return s.repo.Update(fw)
}

func (s *FirmwareService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *FirmwareService) List(page, size int, compatibleType string) ([]model.Firmware, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, compatibleType)
}

// DispatchUpgrade 对指定机器人下发 OTA 升级指令。
func (s *FirmwareService) DispatchUpgrade(robotID, firmwareID string) error {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return fmt.Errorf("robot not found: %w", err)
	}
	if robot.OnlineStatus == 0 {
		return fmt.Errorf("robot is offline")
	}
	fw, err := s.repo.GetByID(firmwareID)
	if err != nil {
		return fmt.Errorf("firmware not found: %w", err)
	}

	return s.publisher.SendConfig(robotID, map[string]interface{}{
		"action":       "upgrade",
		"firmware_id":  fw.FirmwareID,
		"version":      fw.Version,
		"package_path": fw.PackagePath,
	})
}
