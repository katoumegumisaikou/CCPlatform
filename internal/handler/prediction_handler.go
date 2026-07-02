package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PredictionHandler 趋势预测 / 故障预测与健康管理（PHM） HTTP 处理器。
type PredictionHandler struct {
	svc    *service.PredictionService
	phmSvc *service.PHMService
}

func NewPredictionHandler() *PredictionHandler {
	return &PredictionHandler{
		svc:    service.NewPredictionService(),
		phmSvc: service.NewPHMService(),
	}
}

// PredictEfficiency 效率预测接口。
// GET /api/v1/analytics/predictions/efficiency?station_id=xx
func (h *PredictionHandler) PredictEfficiency(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.svc.PredictEfficiency(stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// PredictFault 故障预测接口，支持按机器人、电站或全局维度查询。
// GET /api/v1/analytics/predictions/fault?robot_id=xx 或 ?station_id=xx；不传参数时返回全部活跃预测。
func (h *PredictionHandler) PredictFault(c *gin.Context) {
	robotID := c.Query("robot_id")
	stationID := c.Query("station_id")
	data, err := h.phmSvc.PredictFaults(robotID, stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// GetHealthOverview 设备健康评估接口：查询单台机器人各部件当前健康指数。
// GET /api/v1/analytics/health?robot_id=xx
func (h *PredictionHandler) GetHealthOverview(c *gin.Context) {
	robotID := c.Query("robot_id")
	if robotID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.phmSvc.GetHealthOverview(robotID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// GetHealthByStation 电站维度健康评估接口：查询电站下所有机器人的健康总览。
// GET /api/v1/analytics/health/station?station_id=xx
func (h *PredictionHandler) GetHealthByStation(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.phmSvc.GetHealthByStation(stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// GetHealthTrend 健康趋势分析接口：查询机器人指定部件的历史健康曲线及环比分析。
// GET /api/v1/analytics/health/trend?robot_id=xx&component=drive_motor&days=30
func (h *PredictionHandler) GetHealthTrend(c *gin.Context) {
	robotID := c.Query("robot_id")
	component := c.Query("component")
	if robotID == "" || component == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	days, _ := strconv.Atoi(c.Query("days"))
	data, err := h.phmSvc.GetHealthTrend(robotID, component, days)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// ForecastSpareParts 备件需求预测接口：汇总电站下未处理故障预测涉及的推荐备件。
// GET /api/v1/analytics/predictions/spare-parts?station_id=xx
func (h *PredictionHandler) ForecastSpareParts(c *gin.Context) {
	stationID := c.Query("station_id")
	data, err := h.phmSvc.ForecastSpareParts(stationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, data)
}

// RecommendMaintenanceWindow 维护窗口推荐接口。
// GET /api/v1/analytics/predictions/maintenance-window?robot_id=xx
func (h *PredictionHandler) RecommendMaintenanceWindow(c *gin.Context) {
	robotID := c.Query("robot_id")
	if robotID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.phmSvc.RecommendMaintenanceWindow(robotID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}

// GenerateWorkorder 预测性维护工单生成接口：为指定故障预测手动生成维护工单（若尚未生成）。
// POST /api/v1/analytics/predictions/:id/workorder
func (h *PredictionHandler) GenerateWorkorder(c *gin.Context) {
	predictionID := c.Param("id")
	if predictionID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	handlerName := c.GetString("username")
	maint, err := h.phmSvc.GenerateWorkorder(predictionID, handlerName)
	if err != nil {
		response.Error(c, errcode.ErrPredictionNotFound)
		return
	}
	response.OK(c, maint)
}

// OptimizeStrategy 最优策略推荐接口。
// GET /api/v1/analytics/strategy/optimize?station_id=xx
func (h *PredictionHandler) OptimizeStrategy(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.svc.OptimizeStrategy(stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
}
