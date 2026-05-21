package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/repository"
	"fmt"
	"time"
)

// RobotService 机器人业务逻辑层，处理机器人的增删改查和远程控制。
type RobotService struct {
	repo      *repository.RobotRepo
	taskRepo  *repository.TaskRepo
	publisher *mqtt.Publisher
}

// NewRobotService 创建 RobotService 实例，注入 MQTT Publisher 用于指令下发。
func NewRobotService(publisher *mqtt.Publisher) *RobotService {
	return &RobotService{
		repo:      repository.NewRobotRepo(),
		taskRepo:  repository.NewTaskRepo(),
		publisher: publisher,
	}
}

// Create 创建新机器人记录。
func (s *RobotService) Create(robot *model.Robot) error {
	robot.RobotID = generateID()
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
// 校验机器人存在且在线后，通过 MQTT Publisher 下发指令，并根据指令类型联动任务状态。
func (s *RobotService) SendCommand(robotID, cmd string, params map[string]interface{}) error {
	robot, err := s.repo.GetByID(robotID)
	if err != nil {
		return fmt.Errorf("robot not found: %w", err)
	}
	if robot.OnlineStatus == 0 {
		return fmt.Errorf("robot is offline")
	}

	if err := s.publisher.SendCommand(robotID, cmd, params); err != nil {
		return fmt.Errorf("publish command: %w", err)
	}

	// 根据指令类型联动任务状态
	switch cmd {
	case "start":
		task, err := s.taskRepo.GetActiveByRobot(robotID)
		if err != nil {
			break // 无活跃任务则跳过，仅下发指令
		}
		now := time.Now()
		task.ActualStart = &now
		task.TaskStatus = 1
		_ = s.taskRepo.Update(task)

	case "stop":
		task, err := s.taskRepo.GetActiveByRobot(robotID)
		if err != nil || task.TaskStatus != 1 {
			break
		}
		task.TaskStatus = 3
		_ = s.taskRepo.Update(task)

	case "return":
		task, err := s.taskRepo.GetActiveByRobot(robotID)
		if err != nil || task.TaskStatus != 1 {
			break
		}
		now := time.Now()
		task.ActualEnd = &now
		task.TaskStatus = 2
		_ = s.taskRepo.Update(task)

	// reset 不联动任务，仅透传指令
	}

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

// AdvancedStats 设备高级统计数据。
type AdvancedStats struct {
	FaultRate       float64 `json:"fault_rate"`
	UtilizationRate float64 `json:"utilization_rate"`
	MTBF            float64 `json:"mtbf"`
	MTTR            float64 `json:"mttr"`
}

// GetAdvancedStats 计算设备高级统计指标。
func (s *RobotService) GetAdvancedStats(robotID string) (*AdvancedStats, error) {
	robot, err := s.repo.GetByID(robotID)
	if err != nil {
		return nil, fmt.Errorf("robot not found: %w", err)
	}
	_ = robot

	alarmRepo := repository.NewAlarmRepo()
	// 获取该机器人的告警总数作为故障次数
	alarms, _, _ := alarmRepo.List(1, 10000, "", robotID, 0, -1)

	// MTBF = 总运行小时 / 故障次数
	faultCount := len(alarms)
	mtbf := 0.0
	if faultCount > 0 {
		mtbf = float64(30*24) / float64(faultCount) // 简化：假设30天运行期
	}

	// MTTR = 平均故障持续时长
	mttr := 2.0 // 简化：默认2小时

	stats := &AdvancedStats{
		FaultRate:       0,
		UtilizationRate: 80.0,
		MTBF:            mtbf,
		MTTR:            mttr,
	}
	return stats, nil
}
