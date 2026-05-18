// Package scheduler 提供周期任务调度功能，基于 robfig/cron 实现。
package scheduler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler 周期任务调度器，管理定时和周期清扫任务。
type TaskScheduler struct {
	cron     *cron.Cron
	taskRepo *repository.TaskRepo
}

// NewTaskScheduler 创建调度器实例。
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		cron:     cron.New(cron.WithSeconds()),
		taskRepo: repository.NewTaskRepo(),
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

// ScheduleTask 注册一个周期任务。
func (s *TaskScheduler) ScheduleTask(task *model.Task) error {
	// 移除旧的同任务调度
	s.cron.Remove(cron.EntryID(task.TaskID))

	_, err := s.cron.AddFunc(task.CronExpr, func() {
		newTask := *task
		newTask.TaskID = 0 // 让数据库生成新 ID
		newTask.ActualStart = nil
		newTask.ActualEnd = nil
		newTask.CleanArea = 0
		newTask.TaskStatus = 0 // 待执行
		newTask.CreateTime = time.Now()
		newTask.UpdateTime = time.Now()

		if err := s.taskRepo.Create(&newTask); err != nil {
			log.Printf("[Scheduler] Failed to create periodic task %d: %v", task.TaskID, err)
		} else {
			log.Printf("[Scheduler] Created periodic task %d (from template %d)", newTask.TaskID, task.TaskID)
		}
	})
	return err
}

// RemoveTask 取消周期任务调度。
func (s *TaskScheduler) RemoveTask(taskID uint) {
	s.cron.Remove(cron.EntryID(taskID))
}
