package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"

	"github.com/gin-gonic/gin"
)

// PredictionHandler 趋势预测 HTTP 处理器。
type PredictionHandler struct {
	svc *service.PredictionService
}

func NewPredictionHandler() *PredictionHandler {
	return &PredictionHandler{svc: service.NewPredictionService()}
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

// PredictFault 故障预测接口。
// GET /api/v1/analytics/predictions/fault?robot_id=xx
func (h *PredictionHandler) PredictFault(c *gin.Context) {
	robotID := c.Query("robot_id")
	if robotID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	data, err := h.svc.PredictFaultProbability(robotID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, data)
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
