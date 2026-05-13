package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// TaskRepo 清扫任务数据访问层，封装 tasks 表的所有数据库操作。
type TaskRepo struct {
	db *gorm.DB
}

// NewTaskRepo 创建 TaskRepo 实例，使用全局 DB 连接。
func NewTaskRepo() *TaskRepo {
	return &TaskRepo{db: DB}
}

// Create 创建新任务记录。
func (r *TaskRepo) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

// GetByID 根据任务 ID 查询单条记录。
func (r *TaskRepo) GetByID(id string) (*model.Task, error) {
	var task model.Task
	err := r.db.Where("task_id = ?", id).First(&task).Error
	return &task, err
}

// Update 更新任务信息（全量更新）。
func (r *TaskRepo) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// Delete 根据任务 ID 删除记录。
func (r *TaskRepo) Delete(id string) error {
	return r.db.Where("task_id = ?", id).Delete(&model.Task{}).Error
}

// List 分页查询任务列表，支持按电站、机器人、任务状态筛选。
// taskStatus<0 表示不筛选状态。
func (r *TaskRepo) List(page, size int, stationID, robotID string, taskStatus int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64
	query := r.db.Model(&model.Task{})
	if stationID != "" {
		query = query.Where("station_id = ?", stationID)
	}
	if robotID != "" {
		query = query.Where("robot_id = ?", robotID)
	}
	if taskStatus >= 0 {
		query = query.Where("task_status = ?", taskStatus)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&tasks).Error
	return tasks, total, err
}

// UpdateStatus 更新任务状态，用于任务状态流转（待执行→执行中→已完成等）。
func (r *TaskRepo) UpdateStatus(id string, status int8) error {
	return r.db.Model(&model.Task{}).Where("task_id = ?", id).Update("task_status", status).Error
}

// CountByStation 统计指定电站的任务总数。
func (r *TaskRepo) CountByStation(stationID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Task{}).Where("station_id = ?", stationID).Count(&count).Error
	return count, err
}

// SumCleanAreaByStation 统计指定电站已完成任务的累计清扫面积（㎡）。
// 仅统计 task_status=2（已完成）的任务，用于经济技术指标计算。
func (r *TaskRepo) SumCleanAreaByStation(stationID string) (float64, error) {
	var result struct {
		Total float64
	}
	err := r.db.Model(&model.Task{}).Where("station_id = ? AND task_status = 2", stationID).
		Select("COALESCE(SUM(clean_area), 0) as total").Scan(&result).Error
	return result.Total, err
}

// GetRunningTasks 查询所有执行中的任务（task_status=1），用于监控中心展示。
func (r *TaskRepo) GetRunningTasks() ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("task_status = 1").Find(&tasks).Error
	return tasks, err
}
