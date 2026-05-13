package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// TaskService 清扫任务业务逻辑层，处理任务的创建、状态流转和查询。
type TaskService struct {
	repo      *repository.TaskRepo
	robotRepo *repository.RobotRepo
}

// NewTaskService 创建 TaskService 实例。
func NewTaskService() *TaskService {
	return &TaskService{
		repo:      repository.NewTaskRepo(),
		robotRepo: repository.NewRobotRepo(),
	}
}

// Create 创建新任务。自动从机器人信息中获取电站 ID 并关联。
func (s *TaskService) Create(task *model.Task) error {
	robot, err := s.robotRepo.GetByID(task.RobotID)
	if err != nil {
		return fmt.Errorf("robot not found: %w", err)
	}
	task.StationID = robot.StationID
	return s.repo.Create(task)
}

// GetByID 根据 ID 查询任务详情。
func (s *TaskService) GetByID(id string) (*model.Task, error) {
	return s.repo.GetByID(id)
}

// Update 更新任务信息。
func (s *TaskService) Update(task *model.Task) error {
	return s.repo.Update(task)
}

// Delete 删除任务。
func (s *TaskService) Delete(id string) error {
	return s.repo.Delete(id)
}

// List 分页查询任务列表，支持按电站、机器人、状态筛选。
func (s *TaskService) List(page, size int, stationID, robotID string, taskStatus int) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, stationID, robotID, taskStatus)
}

// UpdateStatus 更新任务状态，包含状态流转校验。
// 状态机:
//
//	0(待执行) → 1(执行中) → 2(已完成)
//	0(待执行) → 3(已暂停) → 4(已取消)
//	1(执行中) → 4(已取消)
func (s *TaskService) UpdateStatus(taskID string, status int8) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	// 状态流转校验
	switch status {
	case 1: // 开始执行：仅待执行状态可启动
		if task.TaskStatus != 0 {
			return fmt.Errorf("task can only start from pending state")
		}
	case 2: // 完成：仅执行中状态可完成
		if task.TaskStatus != 1 {
			return fmt.Errorf("task can only complete from running state")
		}
	case 3: // 暂停：仅执行中状态可暂停
		if task.TaskStatus != 1 {
			return fmt.Errorf("task can only pause from running state")
		}
	case 4: // 取消：已完成的任务不可取消
		if task.TaskStatus == 2 {
			return fmt.Errorf("cannot cancel completed task")
		}
	}
	return s.repo.UpdateStatus(taskID, status)
}
