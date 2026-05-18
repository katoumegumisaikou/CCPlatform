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
func (s *TaskService) GetByID(id uint) (*model.Task, error) {
	return s.repo.GetByID(id)
}

// Update 更新任务信息。
func (s *TaskService) Update(task *model.Task) error {
	return s.repo.Update(task)
}

// Delete 删除任务。
func (s *TaskService) Delete(id uint) error {
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
func (s *TaskService) UpdateStatus(taskID uint, status int8) error {
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

// SmartSchedule 智能调度：选择最优机器人执行任务。
func (s *TaskService) SmartSchedule(stationID string) (*model.Task, error) {
	robots, err := s.robotRepo.GetByStationID(stationID)
	if err != nil || len(robots) == 0 {
		return nil, fmt.Errorf("no robots available in station %s", stationID)
	}
	var best *model.Robot
	for i := range robots {
		r := &robots[i]
		if r.OnlineStatus != 1 || r.BatteryLevel < 20 || r.WorkStatus == 1 {
			continue
		}
		if best == nil || r.BatteryLevel > best.BatteryLevel {
			best = r
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no available robot in station %s", stationID)
	}
	task := &model.Task{
		TaskName:   fmt.Sprintf("智能调度-%s", best.RobotName),
		TaskType:   1,
		RobotID:    best.RobotID,
		StationID:  stationID,
		TaskStatus: 0,
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

// GetProgress 计算任务清扫进度。
func (s *TaskService) GetProgress(taskID uint) (map[string]interface{}, error) {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	result := map[string]interface{}{
		"task_id":     task.TaskID,
		"task_name":   task.TaskName,
		"task_status": task.TaskStatus,
		"clean_area":  task.CleanArea,
	}
	if task.TaskStatus == 1 && task.ActualStart != nil {
		elapsed := 0.0
		if task.PlanEnd != nil && task.PlanStart != nil {
			totalDuration := task.PlanEnd.Sub(*task.PlanStart).Seconds()
			if totalDuration > 0 {
				elapsed = float64(task.UpdateTime.Sub(*task.ActualStart).Seconds())
				progress := elapsed / totalDuration * 100
				if progress > 100 {
					progress = 100
				}
				result["progress"] = progress
				remaining := totalDuration - elapsed
				if remaining < 0 {
					remaining = 0
				}
				result["estimated_remaining_seconds"] = remaining
			}
		}
	}
	if _, ok := result["progress"]; !ok {
		result["progress"] = 0
		result["estimated_remaining_seconds"] = 0
	}
	return result, nil
}
