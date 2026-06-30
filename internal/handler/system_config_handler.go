package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SystemConfigHandler 系统配置 HTTP 处理器。
type SystemConfigHandler struct {
	svc *service.SystemConfigService
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{svc: service.NewSystemConfigService()}
}

// List 分页查询配置列表。
// GET /api/v1/system/configs?page=1&size=10&category=xxx
func (h *SystemConfigHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	category := c.Query("category")

	list, total, err := h.svc.List(page, size, category)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// Get 根据键获取配置。
// GET /api/v1/system/configs/:key
func (h *SystemConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")
	cfg, err := h.svc.GetByKey(key)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, cfg)
}

// Set 设置配置项（键不存在则创建，存在则更新）。
// PUT /api/v1/system/configs/:key
func (h *SystemConfigHandler) Set(c *gin.Context) {
	key := c.Param("key")
	var req model.SystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.ConfigKey = key
	if err := h.svc.Set(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// Delete 删除配置项。
// DELETE /api/v1/system/configs/:key
func (h *SystemConfigHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if err := h.svc.Delete(key); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, nil)
}
