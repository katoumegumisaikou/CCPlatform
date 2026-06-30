package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MaintenanceHandler 维护记录 HTTP 处理器。
type MaintenanceHandler struct {
	svc *service.MaintenanceService
}

func NewMaintenanceHandler() *MaintenanceHandler {
	return &MaintenanceHandler{svc: service.NewMaintenanceService()}
}

// List 分页查询维护记录。
// GET /api/v1/maintenance?page=1&size=10&robot_id=xx&maint_type=xx&start_time=xx&end_time=xx
func (h *MaintenanceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	robotID := c.Query("robot_id")
	maintType := c.Query("maint_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	list, total, err := h.svc.List(page, size, robotID, maintType, startTime, endTime)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// Create 创建维护记录。
// POST /api/v1/maintenance
func (h *MaintenanceHandler) Create(c *gin.Context) {
	var req model.Maintenance
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// Update 更新维护记录。
// PUT /api/v1/maintenance/:id
func (h *MaintenanceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.Maintenance
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.MaintID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, req)
}

// Delete 删除维护记录。
// DELETE /api/v1/maintenance/:id
func (h *MaintenanceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// GetReminders 查询维护提醒。
// GET /api/v1/maintenance/reminders
func (h *MaintenanceHandler) GetReminders(c *gin.Context) {
	list, err := h.svc.GetReminders()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, list)
}
