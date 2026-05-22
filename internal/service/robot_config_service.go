package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/repository"
	"fmt"
)

// RobotConfigService 设备配置模板业务逻辑层。
type RobotConfigService struct {
	repo      *repository.RobotConfigRepo
	robotRepo *repository.RobotRepo
	publisher *mqtt.Publisher
}

func NewRobotConfigService() *RobotConfigService {
	return &RobotConfigService{
		repo:      repository.NewRobotConfigRepo(),
		robotRepo: repository.NewRobotRepo(),
		publisher: mqtt.NewPublisher(mqtt.Server),
	}
}

func (s *RobotConfigService) Create(cfg *model.RobotConfig) error {
	cfg.ConfigID = generateID()
	return s.repo.Create(cfg)
}

func (s *RobotConfigService) Update(cfg *model.RobotConfig) error {
	return s.repo.Update(cfg)
}

func (s *RobotConfigService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *RobotConfigService) List(page, size int, robotType int8) ([]model.RobotConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, robotType)
}

// ApplyConfig 将配置模板应用到指定机器人。
func (s *RobotConfigService) ApplyConfig(robotID, configID string) error {
	robot, err := s.robotRepo.GetByID(robotID)
	if err != nil {
		return fmt.Errorf("robot not found: %w", err)
	}
	if robot.OnlineStatus == 0 {
		return fmt.Errorf("robot is offline")
	}
	cfg, err := s.repo.GetByID(configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	return s.publisher.SendConfig(robotID, map[string]interface{}{
		"action":      "apply_config",
		"config_id":   cfg.ConfigID,
		"config_data": cfg.ConfigData,
	})
}
