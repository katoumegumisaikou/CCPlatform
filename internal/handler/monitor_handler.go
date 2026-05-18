package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MonitorHandler 监控中心 HTTP 处理器，提供仪表盘概览和实时数据查询。
//
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

// GetPositionHistory 查询机器人历史轨迹。
// GET /api/v1/monitor/tracks?robot_id=xx&start_time=xx&end_time=xx&limit=1000
func (h *MonitorHandler) GetPositionHistory(c *gin.Context) {
	robotID := c.Query("robot_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	robotRepo := repository.NewRobotRepo()
	positions, err := robotRepo.GetPositionHistory(robotID, startTime, endTime, limit)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, positions)
}

// GetEnvironmentHistory 查询环境数据历史。
// GET /api/v1/monitor/environment?robot_id=xx&start_time=xx&end_time=xx
func (h *MonitorHandler) GetEnvironmentHistory(c *gin.Context) {
	robotID := c.Query("robot_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	robotRepo := repository.NewRobotRepo()
	envData, err := robotRepo.GetEnvironmentHistory(robotID, startTime, endTime)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, envData)
}

// GetGISConfig 获取 GIS 地图配置。
// GET /api/v1/monitor/gis-config
func (h *MonitorHandler) GetGISConfig(c *gin.Context) {
	cfgRepo := repository.NewSystemConfigRepo()
	cfgs, _ := cfgRepo.GetAllByCategory("gis")
	config := make(map[string]string)
	for _, c := range cfgs {
		config[c.ConfigKey] = c.ConfigValue
	}
	if len(config) == 0 {
		config["map_type"] = "satellite"
		config["default_zoom"] = "12"
		config["default_center"] = "39.9,116.4"
	}
	response.OK(c, config)
}

// GISConfigRequest GIS 配置请求体。
type GISConfigRequest map[string]string

// UpdateGISConfig 更新 GIS 地图配置。
// PUT /api/v1/monitor/gis-config
func (h *MonitorHandler) UpdateGISConfig(c *gin.Context) {
	var req GISConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	cfgRepo := repository.NewSystemConfigRepo()
	for k, v := range req {
		cfg := &model.SystemConfig{
			ConfigKey:   k,
			ConfigValue: v,
			Category:    "gis",
			Description: "GIS地图配置",
		}
		_ = cfgRepo.Set(cfg)
	}
	response.OK(c, gin.H{"message": "updated"})
}
