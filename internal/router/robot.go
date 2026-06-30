package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerRobotRoutes(rg *gin.RouterGroup, h *Handlers) {
	robots := rg.Group("/robots")
	{
		robots.GET("", h.Robot.List)
		robots.POST("", middleware.RBAC("robot:create"), h.Robot.Create)
		robots.GET("/stats", h.Robot.GetStats)
		robots.GET("/:id", h.Robot.GetByID)
		robots.PUT("/:id", middleware.RBAC("robot:update"), h.Robot.Update)
		robots.DELETE("/:id", middleware.RBAC("robot:delete"), h.Robot.Delete)
		robots.POST("/:id/cmd", middleware.RBAC("robot:control"), h.Robot.SendCommand)
		robots.GET("/:id/advanced-stats", h.Robot.GetAdvancedStats)
		robots.POST("/:id/config/apply", middleware.RBAC("robot:config"), h.RobotConfig.ApplyConfig)
	}

	robotConfigs := rg.Group("/robot-configs")
	{
		robotConfigs.GET("", h.RobotConfig.List)
		robotConfigs.POST("", middleware.RBAC("robot:config"), h.RobotConfig.Create)
		robotConfigs.PUT("/:id", middleware.RBAC("robot:config"), h.RobotConfig.Update)
		robotConfigs.DELETE("/:id", middleware.RBAC("robot:config"), h.RobotConfig.Delete)
	}
}
