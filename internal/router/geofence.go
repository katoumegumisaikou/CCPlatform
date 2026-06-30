package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerGeofenceRoutes(rg *gin.RouterGroup, h *Handlers) {
	geofences := rg.Group("/geofences")
	{
		geofences.GET("", h.Geofence.List)
		geofences.GET("/all", h.Geofence.GetAll)
		geofences.POST("", middleware.RBAC("system:config"), h.Geofence.Create)
		geofences.GET("/:id", h.Geofence.GetByID)
		geofences.PUT("/:id", middleware.RBAC("system:config"), h.Geofence.Update)
		geofences.DELETE("/:id", middleware.RBAC("system:config"), h.Geofence.Delete)
	}

	alarms := rg.Group("/geofence-alarms")
	{
		alarms.GET("", h.Geofence.ListAlarms)
		alarms.GET("/recent", h.Geofence.GetRecentAlarms)
		alarms.PUT("/:id", middleware.RBAC("alarm:handle"), h.Geofence.HandleAlarm)
	}
}
