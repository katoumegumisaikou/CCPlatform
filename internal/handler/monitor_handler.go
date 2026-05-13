package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"

	"github.com/gin-gonic/gin"
)

// MonitorHandler 监控中心 HTTP 处理器，提供仪表盘概览和实时数据查询。
type MonitorHandler struct {
	svc *service.MonitorService
}

// NewMonitorHandler 创建 MonitorHandler 实例。
func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{svc: service.NewMonitorService()}
}

// GetDashboard 监控仪表盘接口，聚合电站、机器人、告警等多维度统计数据。
// GET /api/v1/monitor/overview
func (h *MonitorHandler) GetDashboard(c *gin.Context) {
	data, err := h.svc.GetDashboard()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, data)
}

// GetRealtimeData 电站实时数据接口，返回指定电站下所有机器人的位置和状态。
// GET /api/v1/monitor/realtime/:id
func (h *MonitorHandler) GetRealtimeData(c *gin.Context) {
	stationID := c.Param("id")
	data, err := h.svc.GetRealtimeData(stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}
