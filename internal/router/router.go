// Package router 注册所有 HTTP 路由，包括公开接口、JWT 认证接口和 RBAC 权限接口。
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
	systemConfigHandler := handler.NewSystemConfigHandler()
	dictHandler := handler.NewDictHandler()
	orgHandler := handler.NewOrganizationHandler()
	auditHandler := handler.NewAuditHandler()
		firmwareHandler := handler.NewFirmwareHandler()
		maintenanceHandler := handler.NewMaintenanceHandler()
		robotConfigHandler := handler.NewRobotConfigHandler()
	alarmRuleHandler := handler.NewAlarmRuleHandler()
	notifyHandler := handler.NewNotifyHandler()
	cameraHandler := handler.NewCameraHandler()
	predictionHandler := handler.NewPredictionHandler()
	reportTemplateHandler := handler.NewReportTemplateHandler()
	backupHandler := handler.NewBackupHandler()

	v1 := r.Group("/api/v1")
	{
		// Auth (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.GET("/captcha", userHandler.GetCaptcha)
			auth.POST("/verify-captcha", userHandler.VerifyCaptcha)
			auth.POST("/sms-code", userHandler.SendSmsCode)
			auth.POST("/verify-sms", userHandler.VerifySms)
			auth.POST("/forgot-password", userHandler.ForgotPassword)
			auth.POST("/reset-password", userHandler.ResetPassword)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
			protected.Use(middleware.Audit())
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
				robots.POST("", middleware.RBAC("robot:create"), robotHandler.Create)
				robots.GET("/stats", robotHandler.GetStats)
				robots.GET("/:id", robotHandler.GetByID)
				robots.PUT("/:id", middleware.RBAC("robot:update"), robotHandler.Update)
				robots.DELETE("/:id", middleware.RBAC("robot:delete"), robotHandler.Delete)
				robots.POST("/:id/cmd", middleware.RBAC("robot:control"), robotHandler.SendCommand)
					robots.GET("/:id/advanced-stats", robotHandler.GetAdvancedStats)
					robots.POST("/:id/config/apply", middleware.RBAC("robot:config"), robotConfigHandler.ApplyConfig)
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
				analytics.GET("/compare", analyticsHandler.GetComparison)
				analytics.GET("/export", analyticsHandler.ExportReport)
				analytics.GET("/timeseries", analyticsHandler.GetTimeSeries)
				analytics.GET("/predictions/efficiency", predictionHandler.PredictEfficiency)
				analytics.GET("/predictions/fault", predictionHandler.PredictFault)
				analytics.GET("/strategy/optimize", predictionHandler.OptimizeStrategy)
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
					roles.PUT("/:id", middleware.RBAC("role:update"), userHandler.UpdateRole)
					roles.DELETE("/:id", middleware.RBAC("role:delete"), userHandler.DeleteRole)
				}

				// System Configs
				systemConfigs := protected.Group("/system/configs")
				{
					systemConfigs.GET("", systemConfigHandler.List)
					systemConfigs.GET("/:key", systemConfigHandler.Get)
					systemConfigs.PUT("/:key", middleware.RBAC("system:config"), systemConfigHandler.Set)
					systemConfigs.DELETE("/:key", middleware.RBAC("system:config"), systemConfigHandler.Delete)
				}

				// Dicts
				dicts := protected.Group("/dicts")
				{
					dicts.GET("", dictHandler.ListDict)
					dicts.POST("", middleware.RBAC("system:config"), dictHandler.CreateDict)
					dicts.PUT("/:id", middleware.RBAC("system:config"), dictHandler.UpdateDict)
					dicts.DELETE("/:id", middleware.RBAC("system:config"), dictHandler.DeleteDict)
					dicts.GET("/:type/items", dictHandler.ListItems)
					dicts.POST("/:type/items", middleware.RBAC("system:config"), dictHandler.CreateItem)
					dicts.PUT("/items/:id", middleware.RBAC("system:config"), dictHandler.UpdateItem)
					dicts.DELETE("/items/:id", middleware.RBAC("system:config"), dictHandler.DeleteItem)
				}

				// Organizations
				orgs := protected.Group("/organizations")
				{
					orgs.GET("", orgHandler.List)
					orgs.GET("/tree", orgHandler.GetTree)
					orgs.GET("/:id", orgHandler.GetByID)
					orgs.POST("", middleware.RBAC("org:create"), orgHandler.Create)
					orgs.PUT("/:id", middleware.RBAC("org:update"), orgHandler.Update)
					orgs.DELETE("/:id", middleware.RBAC("org:delete"), orgHandler.Delete)
				}

				// Audit Logs
				auditLogs := protected.Group("/audit-logs")
				{
					auditLogs.GET("", auditHandler.List)
				}

				// Login Logs
				loginLogs := protected.Group("/login-logs")
				{
					loginLogs.GET("", userHandler.ListLoginLogs)
				}

				// Firmwares
				firmwares := protected.Group("/firmwares")
				{
					firmwares.GET("", firmwareHandler.List)
					firmwares.POST("", middleware.RBAC("firmware:create"), firmwareHandler.Create)
					firmwares.PUT("/:id", middleware.RBAC("firmware:update"), firmwareHandler.Update)
					firmwares.DELETE("/:id", middleware.RBAC("firmware:delete"), firmwareHandler.Delete)
					firmwares.POST("/:id/upgrade", middleware.RBAC("firmware:upgrade"), firmwareHandler.DispatchUpgrade)
				}

				// Maintenance
				maintenance := protected.Group("/maintenance")
				{
					maintenance.GET("", maintenanceHandler.List)
					maintenance.POST("", middleware.RBAC("maintenance:create"), maintenanceHandler.Create)
					maintenance.PUT("/:id", middleware.RBAC("maintenance:update"), maintenanceHandler.Update)
					maintenance.DELETE("/:id", middleware.RBAC("maintenance:delete"), maintenanceHandler.Delete)
					maintenance.GET("/reminders", maintenanceHandler.GetReminders)
				}

				// Robot Configs
				robotConfigs := protected.Group("/robot-configs")
				{
					robotConfigs.GET("", robotConfigHandler.List)
					robotConfigs.POST("", middleware.RBAC("robot:config"), robotConfigHandler.Create)
					robotConfigs.PUT("/:id", middleware.RBAC("robot:config"), robotConfigHandler.Update)
					robotConfigs.DELETE("/:id", middleware.RBAC("robot:config"), robotConfigHandler.Delete)

				// Alarm Rules
				alarmRules := protected.Group("/alarm-rules")
				{
					alarmRules.GET("", alarmRuleHandler.List)
					alarmRules.POST("", middleware.RBAC("alarm:config"), alarmRuleHandler.Create)
					alarmRules.PUT("/:id", middleware.RBAC("alarm:config"), alarmRuleHandler.Update)
					alarmRules.DELETE("/:id", middleware.RBAC("alarm:config"), alarmRuleHandler.Delete)
				}

				// Notify Templates
				notify := protected.Group("/notify/templates")
				{
					notify.GET("", notifyHandler.List)
					notify.POST("", middleware.RBAC("notify:config"), notifyHandler.Create)
					notify.PUT("/:id", middleware.RBAC("notify:config"), notifyHandler.Update)
					notify.DELETE("/:id", middleware.RBAC("notify:config"), notifyHandler.Delete)
				}

				// Cameras
				cameras := protected.Group("/cameras")
				{
					cameras.GET("", cameraHandler.List)
					cameras.POST("", middleware.RBAC("camera:create"), cameraHandler.Create)
					cameras.PUT("/:id", middleware.RBAC("camera:update"), cameraHandler.Update)
					cameras.DELETE("/:id", middleware.RBAC("camera:delete"), cameraHandler.Delete)
				}

				// Report Templates
				reportTemplates := protected.Group("/report-templates")
				{
					reportTemplates.GET("", reportTemplateHandler.List)
					reportTemplates.POST("", middleware.RBAC("system:config"), reportTemplateHandler.Create)
					reportTemplates.PUT("/:id", middleware.RBAC("system:config"), reportTemplateHandler.Update)
					reportTemplates.DELETE("/:id", middleware.RBAC("system:config"), reportTemplateHandler.Delete)

				// Backups
				backups := protected.Group("/backups")
				{
					backups.GET("", middleware.RBAC("system:config"), backupHandler.List)
					backups.POST("", middleware.RBAC("system:config"), backupHandler.Create)
					backups.DELETE("/:filename", middleware.RBAC("system:config"), backupHandler.Delete)
					backups.POST("/:filename/restore", middleware.RBAC("system:config"), backupHandler.Restore)
				}
				}

				// Robot Tracks & Environment (under robots group is cleaner, but added here for discovery)
				monitor.GET("/tracks", middleware.RBAC("monitor:view"), monitorHandler.GetPositionHistory)
				monitor.GET("/environment", middleware.RBAC("monitor:view"), monitorHandler.GetEnvironmentHistory)
				monitor.GET("/gis-config", monitorHandler.GetGISConfig)
				monitor.PUT("/gis-config", middleware.RBAC("system:config"), monitorHandler.UpdateGISConfig)

				// Task Progress
				tasks.GET("/:id/progress", taskHandler.GetProgress)
				tasks.POST("/smart-schedule", middleware.RBAC("task:create"), taskHandler.SmartSchedule)

				// Alarm Trends
				alarms.GET("/trends", alarmHandler.GetTrendAnalysis)
			}
		}
	}

	return r
}
