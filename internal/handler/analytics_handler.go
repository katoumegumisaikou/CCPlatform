package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler 数据分析 HTTP 处理器，提供经济技术指标和运行报表查询。
type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

// NewAnalyticsHandler 创建 AnalyticsHandler 实例。
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{svc: service.NewAnalyticsService()}
}

// GetEconomicReport 经济效益报表接口，计算电费增收、人工替代率、投资回收期等指标。
// GET /api/v1/analytics/economic?station_id=xxx&start_date=2024-01-01&end_date=2024-12-31
func (h *AnalyticsHandler) GetEconomicReport(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	report, err := h.svc.GetEconomicReport(stationID, start, end)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, report)
}

// GetEfficiencyReport 效率分析报表接口，包含清扫效率、覆盖率、设备利用率等。
// GET /api/v1/analytics/efficiency?station_id=xxx
func (h *AnalyticsHandler) GetEfficiencyReport(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	report, err := h.svc.GetEfficiencyReport(stationID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, report)
}

// GetRunReport 运行报表接口，汇总指定时间段内的清扫面积、任务数、机器人状态等。
// GET /api/v1/analytics/reports?station_id=xxx&start_date=2024-01-01&end_date=2024-12-31
func (h *AnalyticsHandler) GetRunReport(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	report, err := h.svc.GetRunReport(stationID, start, end)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, report)
}
