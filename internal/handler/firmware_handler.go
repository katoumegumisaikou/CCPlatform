package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FirmwareHandler 固件管理 HTTP 处理器。
type FirmwareHandler struct {
	svc *service.FirmwareService
}

func NewFirmwareHandler() *FirmwareHandler {
	return &FirmwareHandler{svc: service.NewFirmwareService()}
}

// List 分页查询固件列表。
// GET /api/v1/firmwares?page=1&size=10&compatible_type=1
func (h *FirmwareHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	compatibleType := c.Query("compatible_type")
	list, total, err := h.svc.List(page, size, compatibleType)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// Create 创建固件记录。
// POST /api/v1/firmwares
func (h *FirmwareHandler) Create(c *gin.Context) {
	var req model.Firmware
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

// Update 更新固件记录。
// PUT /api/v1/firmwares/:id
func (h *FirmwareHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.Firmware
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.FirmwareID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// Delete 删除固件记录。
// DELETE /api/v1/firmwares/:id
func (h *FirmwareHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// UpgradeRobotRequest OTA 升级请求体。
type UpgradeRobotRequest struct {
	RobotIDs []string `json:"robot_ids" binding:"required"`
}

// DispatchUpgrade 批量下发 OTA 升级指令。
// POST /api/v1/firmwares/:id/upgrade
func (h *FirmwareHandler) DispatchUpgrade(c *gin.Context) {
	firmwareID := c.Param("id")
	var req UpgradeRobotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	for _, robotID := range req.RobotIDs {
		if err := h.svc.DispatchUpgrade(robotID, firmwareID); err != nil {
			response.ErrorMsg(c, 400, 10013, err.Error())
			return
		}
	}
	response.OK(c, gin.H{"message": "upgrade dispatched"})
}
