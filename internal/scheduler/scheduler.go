// Package scheduler 提供周期任务调度功能，基于 robfig/cron 实现。
package scheduler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/repository"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler 周期任务调度器，管理定时和周期清扫任务。
type TaskScheduler struct {
	cron      *cron.Cron
	taskRepo  *repository.TaskRepo
	publisher *mqtt.Publisher
}

// NewTaskScheduler 创建调度器实例，注入 MQTT Publisher 用于任务实例下发。
func NewTaskScheduler(publisher *mqtt.Publisher) *TaskScheduler {
	return &TaskScheduler{
		cron:      cron.New(cron.WithSeconds()),
		taskRepo:  repository.NewTaskRepo(),
		publisher: publisher,
	}
}

// Start 启动调度器并加载所有周期任务。
func (s *TaskScheduler) Start() {
	s.cron.Start()
	log.Println("[Scheduler] Task scheduler started")
}

// Stop 停止调度器。
func (s *TaskScheduler) Stop() {
	s.cron.Stop()
	log.Println("[Scheduler] Task scheduler stopped")
}

// ScheduleTask 注册一个周期任务（task_type=3），按 cron 表达式重复执行，不会自动移除。
func (s *TaskScheduler) ScheduleTask(task *model.Task) error {
	s.cron.Remove(cron.EntryID(task.TaskID))

	_, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
	})
	return err
}

// ScheduleOneShot 注册一个一次性定时任务（task_type=2），执行后自动移除 cron。
func (s *TaskScheduler) ScheduleOneShot(task *model.Task) error {
	s.cron.Remove(cron.EntryID(task.TaskID))

	_, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
		s.cron.Remove(cron.EntryID(task.TaskID)) // 执行后移除，避免次年重复触发
		log.Printf("[Scheduler] One-shot task %d completed and removed", task.TaskID)
	})
	return err
}

// executeTask 根据任务模板创建新实例、写入 DB、通过 MQTT 下发给机器人。
func (s *TaskScheduler) executeTask(template *model.Task) {
	newTask := *template
	newTask.TaskID = 0 // 让数据库生成新 ID
	newTask.ActualStart = nil
	newTask.ActualEnd = nil
	newTask.CleanArea = 0
	newTask.TaskStatus = 0 // 待执行
	newTask.CreateTime = time.Now()
	newTask.UpdateTime = time.Now()

	if err := s.taskRepo.Create(&newTask); err != nil {
		log.Printf("[Scheduler] Failed to create task from template %d: %v", template.TaskID, err)
		return
	}
	log.Printf("[Scheduler] Created task %d (from template %d)", newTask.TaskID, template.TaskID)

	// 下发新任务实例到机器人
	params := map[string]interface{}{
		"task_id":   newTask.TaskID,
		"task_type": newTask.TaskType,
		"area_ids":  newTask.AreaIDs,
	}
	if newTask.PlanStart != nil {
		params["plan_start"] = newTask.PlanStart.Format("2006-01-02 15:04:05")
	}
	if newTask.PlanEnd != nil {
		params["plan_end"] = newTask.PlanEnd.Format("2006-01-02 15:04:05")
	}
	_ = s.publisher.SendCommand(newTask.RobotID, "task", params)
}

// RemoveTask 取消周期任务调度。
func (s *TaskScheduler) RemoveTask(taskID uint) {
	s.cron.Remove(cron.EntryID(taskID))
}
