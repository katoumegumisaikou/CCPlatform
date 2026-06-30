package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/scheduler"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// TaskHandler 清扫任务 HTTP 处理器，处理任务的创建、查询和状态变更。
//
type TaskHandler struct {
	svc *service.TaskService
}

// NewTaskHandler 创建 TaskHandler 实例，publisher 和 scheduler 透传给 TaskService。
func NewTaskHandler(publisher *mqtt.Publisher, sch *scheduler.TaskScheduler) *TaskHandler {
	return &TaskHandler{svc: service.NewTaskService(publisher, sch)}
}

// List 任务列表接口，支持分页和多维度筛选。
// GET /api/v1/tasks?page=1&size=10&station_id=xxx&robot_id=xxx&task_status=1
func (h *TaskHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	stationID := c.Query("station_id")
	robotID := c.Query("robot_id")
	taskStatus, _ := strconv.Atoi(c.DefaultQuery("task_status", "-1"))

	tasks, total, err := h.svc.List(page, size, stationID, robotID, taskStatus)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, tasks, total, page, size)
}

// GetByID 任务详情接口。
// GET /api/v1/tasks/:id
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	task, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, task)
}

// CreateTaskRequest 创建任务请求体。
type CreateTaskRequest struct {
	TaskName  string     `json:"task_name" binding:"required"` // 任务名称（必填）
	TaskType  int8       `json:"task_type" binding:"required"` // 任务类型(1即时 2定时 3周期)
	RobotID   string     `json:"robot_id" binding:"required"`  // 执行机器人 ID（必填）
	AreaIDs   string     `json:"area_ids"`                     // 清扫区域 ID 列表
	PlanStart *time.Time `json:"plan_start"`                   // 计划开始时间
	PlanEnd   *time.Time `json:"plan_end"`                     // 计划结束时间
	CronExpr  string     `json:"cron_expr"`                    // 周期任务 cron 表达式（task_type=3 时必填，如 "0 0 8 * * *"）
}

// Create 创建任务接口，自动关联机器人所属电站。
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	task := &model.Task{
		TaskName:   req.TaskName,
		TaskType:   req.TaskType,
		RobotID:    req.RobotID,
		AreaIDs:    req.AreaIDs,
		PlanStart:  req.PlanStart,
		PlanEnd:    req.PlanEnd,
		CronExpr:   req.CronExpr,
		TaskStatus: 0,
	}
	if err := h.svc.Create(task); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, task)
}

// UpdateTaskStatusRequest 更新任务状态请求体。
type UpdateTaskStatusRequest struct {
	Status int8 `json:"status" binding:"required"` // 目标状态(1执行 2完成 3暂停 4取消)
}

// UpdateStatus 任务状态变更接口，包含状态机校验。
// PUT /api/v1/tasks/:id/status
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.UpdateStatus(uint(id), req.Status); err != nil {
		response.ErrorMsg(c, 400, 10010, err.Error())
		return
	}
	response.OK(c, nil)
}

// Delete 删除任务接口。
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// SmartScheduleRequest 智能调度请求体。
type SmartScheduleRequest struct {
	StationID string `json:"station_id" binding:"required"`
}

// SmartSchedule 智能调度接口，选择最优机器人创建任务。
// POST /api/v1/tasks/smart-schedule
func (h *TaskHandler) SmartSchedule(c *gin.Context) {
	var req SmartScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	task, err := h.svc.SmartSchedule(req.StationID)
	if err != nil {
		response.ErrorMsg(c, 400, 10018, err.Error())
		return
	}
	response.OK(c, task)
}

// GetProgress 任务清扫进度查询接口。
// GET /api/v1/tasks/:id/progress
func (h *TaskHandler) GetProgress(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	data, err := h.svc.GetProgress(uint(id))
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}
