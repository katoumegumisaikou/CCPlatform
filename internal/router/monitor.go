package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerMonitorRoutes(rg *gin.RouterGroup, h *Handlers) {
	monitor := rg.Group("/monitor")
	{
		monitor.GET("/overview", h.Monitor.GetDashboard)
		monitor.GET("/tracks", middleware.RBAC("monitor:view"), h.Monitor.GetPositionHistory)
		monitor.GET("/environment", middleware.RBAC("monitor:view"), h.Monitor.GetEnvironmentHistory)
		monitor.GET("/gis-config", h.Monitor.GetGISConfig)
		monitor.PUT("/gis-config", middleware.RBAC("system:config"), h.Monitor.UpdateGISConfig)
	}
}
