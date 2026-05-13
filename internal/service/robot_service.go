package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// RobotService 机器人业务逻辑层，处理机器人的增删改查和远程控制。
type RobotService struct {
	repo     *repository.RobotRepo
	taskRepo *repository.TaskRepo
}

// NewRobotService 创建 RobotService 实例。
func NewRobotService() *RobotService {
	return &RobotService{
		repo:     repository.NewRobotRepo(),
		taskRepo: repository.NewTaskRepo(),
	}
}

// Create 创建新机器人记录。
func (s *RobotService) Create(robot *model.Robot) error {
	return s.repo.Create(robot)
}

// GetByID 根据 ID 查询机器人详情。
func (s *RobotService) GetByID(id string) (*model.Robot, error) {
	return s.repo.GetByID(id)
}

// Update 更新机器人信息。
func (s *RobotService) Update(robot *model.Robot) error {
	return s.repo.Update(robot)
}

// Delete 删除机器人。
func (s *RobotService) Delete(id string) error {
	return s.repo.Delete(id)
}

// List 分页查询机器人列表，支持按电站、类型、在线状态筛选。
func (s *RobotService) List(page, size int, stationID string, robotType, onlineStatus int) ([]model.Robot, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, stationID, robotType, onlineStatus)
}

// SendCommand 向机器人下发控制指令。
// 先校验机器人是否存在且在线，校验通过后由 Handler 层通过 MQTT Publisher 发送。
func (s *RobotService) SendCommand(robotID, cmd string, params map[string]interface{}) error {
	robot, err := s.repo.GetByID(robotID)
	if err != nil {
		return fmt.Errorf("robot not found: %w", err)
	}
	if robot.OnlineStatus == 0 {
		return fmt.Errorf("robot is offline")
	}
	// MQTT client will handle actual publishing
	return nil
}

// GetStats 获取机器人状态统计数据（总数/在线/离线/故障），用于仪表盘。
func (s *RobotService) GetStats() (map[string]int64, error) {
	return s.repo.CountByStatus()
}

// GetByStationID 查询指定电站下的所有机器人。
func (s *RobotService) GetByStationID(stationID string) ([]model.Robot, error) {
	return s.repo.GetByStationID(stationID)
}
