package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ReportTemplateHandler 报表模板 HTTP 处理器。
type ReportTemplateHandler struct {
	svc *service.ReportTemplateService
}

func NewReportTemplateHandler() *ReportTemplateHandler {
	return &ReportTemplateHandler{svc: service.NewReportTemplateService()}
}

func (h *ReportTemplateHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.svc.List(page, size)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

func (h *ReportTemplateHandler) Create(c *gin.Context) {
	var req model.ReportTemplate
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

func (h *ReportTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.ReportTemplate
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

func (h *ReportTemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
