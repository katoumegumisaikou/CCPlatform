package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerWeatherRoutes 注册气象数据相关路由。
// 查看权限对所有已登录用户开放，管理权限 (weather:manage) 仅对系统管理员和电站管理员开放。
func registerWeatherRoutes(rg *gin.RouterGroup, h *Handlers) {
	weather := rg.Group("/weather")
	{
		// 实时天气
		weather.GET("/now", h.Weather.GetNowWeather)
		weather.POST("/refresh", middleware.RBAC("weather:manage"), h.Weather.RefreshWeather)

		// 预报
		weather.GET("/forecast/daily", h.Weather.GetDailyForecast)
		weather.GET("/forecast/hourly", h.Weather.GetHourlyForecast)

		// 预警
		weather.GET("/warnings", h.Weather.GetActiveWarnings)
		weather.GET("/warnings/history", h.Weather.GetWarningHistory)

		// 空气质量
		weather.GET("/air-quality", h.Weather.GetAirQuality)
	}
}

// registerCleaningDecisionRoutes 注册清扫决策相关路由。
func registerCleaningDecisionRoutes(rg *gin.RouterGroup, h *Handlers) {
	cd := rg.Group("/cleaning-decisions")
	{
		// 清扫决策规则
		cd.GET("/rules", h.CleaningDecision.ListRules)
		cd.POST("/rules", middleware.RBAC("weather:manage"), h.CleaningDecision.CreateRule)
		cd.GET("/rules/:id", h.CleaningDecision.GetRule)
		cd.PUT("/rules/:id", middleware.RBAC("weather:manage"), h.CleaningDecision.UpdateRule)
		cd.DELETE("/rules/:id", middleware.RBAC("weather:manage"), h.CleaningDecision.DeleteRule)

		// 极端天气规则
		cd.GET("/extreme-rules", h.CleaningDecision.ListExtremeRules)
		cd.POST("/extreme-rules", middleware.RBAC("weather:manage"), h.CleaningDecision.CreateExtremeRule)
		cd.GET("/extreme-rules/:id", h.CleaningDecision.GetExtremeRule)
		cd.PUT("/extreme-rules/:id", middleware.RBAC("weather:manage"), h.CleaningDecision.UpdateExtremeRule)
		cd.DELETE("/extreme-rules/:id", middleware.RBAC("weather:manage"), h.CleaningDecision.DeleteExtremeRule)

		// 清扫建议
		cd.GET("/advice", h.CleaningDecision.GetCleaningAdvice)

		// 动态频次调整
		cd.GET("/adjustments", h.CleaningDecision.GetAdjustments)
		cd.POST("/adjustments", middleware.RBAC("weather:manage"), h.CleaningDecision.GenerateAdjustment)
	}
}
