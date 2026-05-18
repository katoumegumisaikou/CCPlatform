package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler 数据分析 HTTP 处理器，提供经济技术指标和运行报表查询。
//
// TODO: 真实效率计算 — 覆盖率/重复率应从实际清扫轨迹数据计算，替换硬编码常量 (需求 4.5.2)
// TODO: 对比分析 — 多电站/多机器人数据对比 API (需求 4.5.2)
// TODO: 趋势预测 — 发电效率预测模型、故障预测模型、最优清扫策略模型 (需求 8.2)
// TODO: 自定义报表 — 用户自定义报表模板和数据维度配置 (需求 4.5.2)
// TODO: 报表导出 — 支持 Excel/PDF 格式导出 (需求 4.5.2)
// TODO: 时间维度聚合 — 按日/周/月/年聚合的查询接口 (需求 4.5.2)
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
