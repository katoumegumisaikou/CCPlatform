package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RobotConfigHandler 设备配置模板 HTTP 处理器。
type RobotConfigHandler struct {
	svc *service.RobotConfigService
}

func NewRobotConfigHandler() *RobotConfigHandler {
	return &RobotConfigHandler{svc: service.NewRobotConfigService()}
}

// List 分页查询配置模板。
// GET /api/v1/robot-configs?robot_type=1&page=1&size=10
func (h *RobotConfigHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	robotType, _ := strconv.Atoi(c.DefaultQuery("robot_type", "0"))
	list, total, err := h.svc.List(page, size, int8(robotType))
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// Create 创建配置模板。
// POST /api/v1/robot-configs
func (h *RobotConfigHandler) Create(c *gin.Context) {
	var req model.RobotConfig
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

// Update 更新配置模板。
// PUT /api/v1/robot-configs/:id
func (h *RobotConfigHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.RobotConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.ConfigID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// Delete 删除配置模板。
// DELETE /api/v1/robot-configs/:id
func (h *RobotConfigHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// ApplyConfigRequest 应用配置请求体。
type ApplyConfigRequest struct {
	ConfigID string `json:"config_id" binding:"required"`
}

// ApplyConfig 将配置模板应用到指定机器人。
// POST /api/v1/robots/:id/config/apply
func (h *RobotConfigHandler) ApplyConfig(c *gin.Context) {
	robotID := c.Param("id")
	var req ApplyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.ApplyConfig(robotID, req.ConfigID); err != nil {
		response.ErrorMsg(c, 400, 10009, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "config applied"})
}
