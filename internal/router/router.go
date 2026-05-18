// Package router 注册所有 HTTP 路由，包括公开接口、JWT 认证接口和 RBAC 权限接口。
//
// TODO: 设备注册路由 — POST /api/v1/robots (需求 4.3)
// TODO: 设备更新/删除路由 — PUT/DELETE /api/v1/robots/:id (需求 4.3)
// TODO: OTA 固件升级路由 — /api/v1/firmware/* (需求 4.3)
// TODO: 维护管理路由 — /api/v1/maintenance/* (需求 4.3)
// TODO: 告警规则路由 — /api/v1/alarm-rules/* (需求 4.4.2)
// TODO: 告警通知配置路由 — /api/v1/notify/* (需求 4.4.2)
// TODO: 组织架构路由 — /api/v1/organizations/* (需求 4.6.2)
// TODO: 操作日志路由 — /api/v1/audit-logs (需求 4.6.2)
// TODO: 系统配置/字典路由 — /api/v1/system/* (需求 4.6.2)
// TODO: 数据备份路由 — /api/v1/backup/* (需求 4.6.2)
// TODO: 报表导出路由 — /api/v1/analytics/export (需求 4.5.2)
// TODO: 对比分析路由 — /api/v1/analytics/compare (需求 4.5.2)
// TODO: 历史轨迹路由 — /api/v1/robots/:id/tracks (需求 4.1.2)
// TODO: 视频监控路由 — /api/v1/cameras/* (需求 4.1.2)
package router

import (
	"ccplatform/internal/handler"
	"ccplatform/internal/middleware"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/ws"

	"github.com/gin-gonic/gin"
)

func Setup(hub *ws.Hub, publisher *mqtt.Publisher) *gin.Engine {
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS())

	// WebSocket endpoint
	r.GET("/ws", ws.HandleWebSocket(hub))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Handlers
	userHandler := handler.NewUserHandler()
	stationHandler := handler.NewStationHandler()
	robotHandler := handler.NewRobotHandler(publisher)
	taskHandler := handler.NewTaskHandler()
	alarmHandler := handler.NewAlarmHandler()
	monitorHandler := handler.NewMonitorHandler()
	analyticsHandler := handler.NewAnalyticsHandler()

	v1 := r.Group("/api/v1")
	{
		// Auth (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/auth/me", userHandler.GetCurrentUser)

			// Stations
			stations := protected.Group("/stations")
			{
				stations.GET("", stationHandler.List)
				stations.POST("", middleware.RBAC("station:create"), stationHandler.Create)
				stations.GET("/all", stationHandler.GetAll)
				stations.GET("/:id", stationHandler.GetByID)
				stations.PUT("/:id", middleware.RBAC("station:update"), stationHandler.Update)
				stations.DELETE("/:id", middleware.RBAC("station:delete"), stationHandler.Delete)
				stations.GET("/:id/robots", robotHandler.GetByStation)
				stations.GET("/:id/realtime", monitorHandler.GetRealtimeData)
			}

			// Robots
			robots := protected.Group("/robots")
			{
				robots.GET("", robotHandler.List)
				robots.GET("/stats", robotHandler.GetStats)
				robots.GET("/:id", robotHandler.GetByID)
				robots.POST("/:id/cmd", middleware.RBAC("robot:control"), robotHandler.SendCommand)
			}

			// Tasks
			tasks := protected.Group("/tasks")
			{
				tasks.GET("", taskHandler.List)
				tasks.POST("", middleware.RBAC("task:create"), taskHandler.Create)
				tasks.GET("/:id", taskHandler.GetByID)
				tasks.PUT("/:id/status", middleware.RBAC("task:update"), taskHandler.UpdateStatus)
				tasks.DELETE("/:id", middleware.RBAC("task:delete"), taskHandler.Delete)
			}

			// Alarms
			alarms := protected.Group("/alarms")
			{
				alarms.GET("", alarmHandler.List)
				alarms.GET("/stats", alarmHandler.GetStats)
				alarms.GET("/recent", alarmHandler.GetRecent)
				alarms.GET("/:id", alarmHandler.GetByID)
				alarms.PUT("/:id", middleware.RBAC("alarm:handle"), alarmHandler.Handle)
			}

			// Monitor
			monitor := protected.Group("/monitor")
			{
				monitor.GET("/overview", monitorHandler.GetDashboard)
			}

			// Analytics
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/economic", analyticsHandler.GetEconomicReport)
				analytics.GET("/efficiency", analyticsHandler.GetEfficiencyReport)
				analytics.GET("/reports", analyticsHandler.GetRunReport)
			}

			// Users
			users := protected.Group("/users")
			{
				users.GET("", middleware.RBAC("user:view"), userHandler.List)
				users.POST("", middleware.RBAC("user:create"), userHandler.Create)
				users.PUT("/:id", middleware.RBAC("user:update"), userHandler.Update)
				users.DELETE("/:id", middleware.RBAC("user:delete"), userHandler.Delete)
			}

			// Roles
			roles := protected.Group("/roles")
			{
				roles.GET("", userHandler.ListRoles)
				roles.POST("", middleware.RBAC("role:create"), userHandler.CreateRole)
			}
		}
	}

	return r
}
