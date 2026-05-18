package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AlarmRuleHandler 告警规则 HTTP 处理器。
type AlarmRuleHandler struct {
	svc *service.AlarmRuleService
}

func NewAlarmRuleHandler() *AlarmRuleHandler {
	return &AlarmRuleHandler{svc: service.NewAlarmRuleService()}
}

func (h *AlarmRuleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	alarmType := c.Query("alarm_type")
	list, total, err := h.svc.List(page, size, alarmType)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

func (h *AlarmRuleHandler) Create(c *gin.Context) {
	var req model.AlarmRule
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

func (h *AlarmRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.AlarmRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.RuleID = id
	if err := h.svc.Update(&req); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, req)
}

func (h *AlarmRuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
