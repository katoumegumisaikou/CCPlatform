package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotifyHandler 通知模板 HTTP 处理器。
type NotifyHandler struct {
	svc *service.NotifyService
}

func NewNotifyHandler() *NotifyHandler {
	return &NotifyHandler{svc: service.NewNotifyService()}
}

func (h *NotifyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	tplType := c.Query("type")
	list, total, err := h.svc.List(page, size, tplType)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

func (h *NotifyHandler) Create(c *gin.Context) {
	var req model.NotifyTemplate
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

func (h *NotifyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.NotifyTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.TplID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

func (h *NotifyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
