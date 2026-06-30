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
		analytics.GET("/strategy/optimize", h.Prediction.OptimizeStrategy)
	}
}
