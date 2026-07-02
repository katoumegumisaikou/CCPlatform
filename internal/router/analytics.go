package router

import "github.com/gin-gonic/gin"

func registerAnalyticsRoutes(rg *gin.RouterGroup, h *Handlers) {
	analytics := rg.Group("/analytics")
	{
		analytics.GET("/economic", h.Analytics.GetEconomicReport)
		analytics.GET("/efficiency", h.Analytics.GetEfficiencyReport)
		analytics.GET("/reports", h.Analytics.GetRunReport)
		analytics.GET("/compare", h.Analytics.GetComparison)
		analytics.GET("/export", h.Analytics.ExportReport)
		analytics.GET("/timeseries", h.Analytics.GetTimeSeries)
		analytics.GET("/predictions/efficiency", h.Prediction.PredictEfficiency)
		analytics.GET("/predictions/fault", h.Prediction.PredictFault)
		analytics.GET("/predictions/spare-parts", h.Prediction.ForecastSpareParts)
		analytics.GET("/predictions/maintenance-window", h.Prediction.RecommendMaintenanceWindow)
		analytics.POST("/predictions/:id/workorder", h.Prediction.GenerateWorkorder)
		analytics.GET("/strategy/optimize", h.Prediction.OptimizeStrategy)
		analytics.GET("/health", h.Prediction.GetHealthOverview)
		analytics.GET("/health/station", h.Prediction.GetHealthByStation)
		analytics.GET("/health/trend", h.Prediction.GetHealthTrend)
	}
}
