package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeatherHandler 气象 HTTP 处理器，处理气象数据相关的 RESTful API 请求。
type WeatherHandler struct {
	svc *service.WeatherService
}

// NewWeatherHandler 创建 WeatherHandler 实例。
func NewWeatherHandler() *WeatherHandler {
	return &WeatherHandler{svc: service.NewWeatherService()}
}

// RefreshStationWeatherRequest 手动刷新天气请求体。
type RefreshStationWeatherRequest struct {
	StationID string `json:"station_id" binding:"required"` // 电站 ID
}

// ─── 实时天气 ─────────────────────────────────────────────────

// GetNowWeather 获取电站最新实时天气。
// GET /api/v1/weather/now?station_id=xxx
func (h *WeatherHandler) GetNowWeather(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}

	now, err := h.svc.GetNowWeather(stationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, errcode.ErrNotFound)
			return
		}
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, now)
}

// RefreshWeather 手动触发天气数据刷新。
// POST /api/v1/weather/refresh
func (h *WeatherHandler) RefreshWeather(c *gin.Context) {
	var req RefreshStationWeatherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}

	if err := h.svc.RefreshStationWeather(req.StationID); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// ─── 预报 ─────────────────────────────────────────────────────

// GetDailyForecast 获取电站逐天预报。
// GET /api/v1/weather/forecast/daily?station_id=xxx&days=7
func (h *WeatherHandler) GetDailyForecast(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	list, err := h.svc.GetDailyForecast(stationID, days)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, list)
}

// GetHourlyForecast 获取电站逐小时预报。
// GET /api/v1/weather/forecast/hourly?station_id=xxx&hours=24
func (h *WeatherHandler) GetHourlyForecast(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	list, err := h.svc.GetHourlyForecast(stationID, hours)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, list)
}

// ─── 预警 ─────────────────────────────────────────────────────

// GetActiveWarnings 获取电站当前生效的天气预警。
// GET /api/v1/weather/warnings?station_id=xxx
func (h *WeatherHandler) GetActiveWarnings(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}

	list, err := h.svc.GetActiveWarnings(stationID)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, list)
}

// GetWarningHistory 分页查询历史预警记录。
// GET /api/v1/weather/warnings/history?station_id=xxx&page=1&size=10&start=2024-01-01&end=2024-12-31
func (h *WeatherHandler) GetWarningHistory(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	startDate := c.Query("start")
	endDate := c.Query("end")

	list, total, err := h.svc.GetWarningHistory(stationID, page, size, startDate, endDate)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// ─── 空气质量 ─────────────────────────────────────────────────

// GetAirQuality 获取电站最新空气质量数据。
// GET /api/v1/weather/air-quality?station_id=xxx
func (h *WeatherHandler) GetAirQuality(c *gin.Context) {
	stationID := c.Query("station_id")
	if stationID == "" {
		response.Error(c, errcode.ErrParam)
		return
	}

	air, err := h.svc.GetAirQuality(stationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, errcode.ErrNotFound)
			return
		}
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, air)
}
