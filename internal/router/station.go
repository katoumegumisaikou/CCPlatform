package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerStationRoutes(rg *gin.RouterGroup, h *Handlers) {
	stations := rg.Group("/stations")
	{
		stations.GET("", h.Station.List)
		stations.POST("", middleware.RBAC("station:create"), h.Station.Create)
		stations.GET("/all", h.Station.GetAll)
		stations.GET("/:id", h.Station.GetByID)
		stations.PUT("/:id", middleware.RBAC("station:update"), h.Station.Update)
		stations.DELETE("/:id", middleware.RBAC("station:delete"), h.Station.Delete)
		stations.GET("/:id/robots", h.Robot.GetByStation)
		stations.GET("/:id/realtime", h.Monitor.GetRealtimeData)
	}
}
