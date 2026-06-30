package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerDeviceRoutes(rg *gin.RouterGroup, h *Handlers) {
	firmwares := rg.Group("/firmwares")
	{
		firmwares.GET("", h.Firmware.List)
		firmwares.POST("", middleware.RBAC("firmware:create"), h.Firmware.Create)
		firmwares.PUT("/:id", middleware.RBAC("firmware:update"), h.Firmware.Update)
		firmwares.DELETE("/:id", middleware.RBAC("firmware:delete"), h.Firmware.Delete)
		firmwares.POST("/:id/upgrade", middleware.RBAC("firmware:upgrade"), h.Firmware.DispatchUpgrade)
	}

	maintenance := rg.Group("/maintenance")
	{
		maintenance.GET("", h.Maintenance.List)
		maintenance.POST("", middleware.RBAC("maintenance:create"), h.Maintenance.Create)
		maintenance.PUT("/:id", middleware.RBAC("maintenance:update"), h.Maintenance.Update)
		maintenance.DELETE("/:id", middleware.RBAC("maintenance:delete"), h.Maintenance.Delete)
		maintenance.GET("/reminders", h.Maintenance.GetReminders)
	}

	cameras := rg.Group("/cameras")
	{
		cameras.GET("", h.Camera.List)
		cameras.POST("", middleware.RBAC("camera:create"), h.Camera.Create)
		cameras.PUT("/:id", middleware.RBAC("camera:update"), h.Camera.Update)
		cameras.DELETE("/:id", middleware.RBAC("camera:delete"), h.Camera.Delete)
	}
}
