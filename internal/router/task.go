package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerTaskRoutes(rg *gin.RouterGroup, h *Handlers) {
	tasks := rg.Group("/tasks")
	{
		tasks.GET("", h.Task.List)
		tasks.POST("", middleware.RBAC("task:create"), h.Task.Create)
		tasks.GET("/:id", h.Task.GetByID)
		tasks.PUT("/:id/status", middleware.RBAC("task:update"), h.Task.UpdateStatus)
		tasks.DELETE("/:id", middleware.RBAC("task:delete"), h.Task.Delete)
		tasks.GET("/:id/progress", h.Task.GetProgress)
		tasks.POST("/smart-schedule", middleware.RBAC("task:create"), h.Task.SmartSchedule)
	}
}
