package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AlarmHandler 告警 HTTP 处理器，处理告警的查询、处理和统计。
//
type AlarmHandler struct {
	svc *service.AlarmService
}

// NewAlarmHandler 创建 AlarmHandler 实例。
func NewAlarmHandler() *AlarmHandler {
	return &AlarmHandler{svc: service.NewAlarmService()}
}

// List 告警列表接口，支持分页和多维度筛选。
// GET /api/v1/alarms?page=1&size=10&station_id=xxx&robot_id=xxx&alarm_level=3&handle_status=0
func (h *AlarmHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	stationID := c.Query("station_id")
	robotID := c.Query("robot_id")
	alarmLevel, _ := strconv.Atoi(c.DefaultQuery("alarm_level", "0"))
	handleStatus, _ := strconv.Atoi(c.DefaultQuery("handle_status", "-1"))

	alarms, total, err := h.svc.List(page, size, stationID, robotID, alarmLevel, handleStatus)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, alarms, total, page, size)
}

// GetByID 告警详情接口。
// GET /api/v1/alarms/:id
func (h *AlarmHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	alarm, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, alarm)
}

// HandleAlarmRequest 处理告警请求体。
type HandleAlarmRequest struct {
	HandleStatus int8 `json:"handle_status" binding:"required"` // 处理状态(1已确认 2已处理 3已忽略)
}

// Handle 处理告警接口，记录处理人（从 JWT Token 中获取）和处理状态。
// PUT /api/v1/alarms/:id
func (h *AlarmHandler) Handle(c *gin.Context) {
	id := c.Param("id")
	var req HandleAlarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	handlerID := c.GetString("user_id") // 从 JWT 中间件注入的用户 ID
	if err := h.svc.Handle(id, handlerID, req.HandleStatus); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// GetStats 告警统计接口，按级别分组统计未处理告警。
// GET /api/v1/alarms/stats
func (h *AlarmHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, stats)
}

// GetRecent 最近告警接口，用于监控中心实时告警展示。
// GET /api/v1/alarms/recent?limit=10
func (h *AlarmHandler) GetRecent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	alarms, err := h.svc.GetRecentAlarms(limit)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, alarms)
	}

	// GetTrendAnalysis 告警趋势分析接口。
	// GET /api/v1/alarms/trends?granularity=day&start_time=xx&end_time=xx
	func (h *AlarmHandler) GetTrendAnalysis(c *gin.Context) {
		granularity := c.DefaultQuery("granularity", "day")
		startTime := c.Query("start_time")
		endTime := c.Query("end_time")
		data, err := h.svc.GetTrendAnalysis(granularity, startTime, endTime)
		if err != nil {
			response.Error(c, errcode.ErrInternal)
			return
		}
		response.OK(c, data)
	}
