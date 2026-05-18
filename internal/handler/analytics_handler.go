package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler 数据分析 HTTP 处理器，提供经济技术指标和运行报表查询。
//
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

// GetComparison 多维度对比分析接口。
// GET /api/v1/analytics/compare?dimension=station&ids=id1,id2
func (h *AnalyticsHandler) GetComparison(c *gin.Context) {
	dimension := c.DefaultQuery("dimension", "station")
	idsStr := c.Query("ids")
	if idsStr == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	ids := make([]string, 0)
	for _, id := range strings.Split(idsStr, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	data, err := h.svc.Compare(dimension, ids)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, data)
}

// ExportReport 报表导出接口，返回 CSV 格式。
// GET /api/v1/analytics/export?station_id=xx&start_date=xx&end_date=xx
func (h *AnalyticsHandler) ExportReport(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	rows, err := h.svc.ExportReport(stationID, start, end)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var sb strings.Builder
	for _, row := range rows {
		sb.WriteString(strings.Join(row, ","))
		sb.WriteString("\n")
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.csv", stationID))
	c.String(200, sb.String())
}

// GetTimeSeries 时间维度聚合接口。
// GET /api/v1/analytics/timeseries?station_id=xx&granularity=day&start_time=xx&end_time=xx
func (h *AnalyticsHandler) GetTimeSeries(c *gin.Context) {
	stationID := c.Query("station_id")
	granularity := c.DefaultQuery("granularity", "day")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	data, err := h.svc.GetTimeSeries(stationID, granularity, startTime, endTime)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, data)
}
