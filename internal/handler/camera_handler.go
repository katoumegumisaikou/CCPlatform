package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CameraHandler 摄像头管理 HTTP 处理器。
type CameraHandler struct {
	svc *service.CameraService
}

func NewCameraHandler() *CameraHandler {
	return &CameraHandler{svc: service.NewCameraService()}
}

func (h *CameraHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	stationID := c.Query("station_id")
	list, total, err := h.svc.List(page, size, stationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

func (h *CameraHandler) Create(c *gin.Context) {
	var req model.Camera
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

func (h *CameraHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.Camera
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.CameraID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

func (h *CameraHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
