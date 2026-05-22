// Package router 注册所有 HTTP 路由，包括公开接口、JWT 认证接口和 RBAC 权限接口。
// 路由按功能模块拆分：auth / station / robot / task / alarm / monitor / analytics /
// user / system / organization / device / report，每个模块在独立文件中注册。
package router

import (
	"ccplatform/internal/handler"
	"ccplatform/internal/middleware"
	"ccplatform/internal/mqtt"
	"ccplatform/internal/scheduler"
	"ccplatform/internal/ws"

	"github.com/gin-gonic/gin"
)

// Handlers 聚合所有 HTTP handler 实例，供各路由注册函数使用。
type Handlers struct {
	User           *handler.UserHandler
	Station        *handler.StationHandler
	Robot          *handler.RobotHandler
	Task           *handler.TaskHandler
	Alarm          *handler.AlarmHandler
	Monitor        *handler.MonitorHandler
	Analytics      *handler.AnalyticsHandler
	SystemConfig   *handler.SystemConfigHandler
	Dict           *handler.DictHandler
	Organization   *handler.OrganizationHandler
	Audit          *handler.AuditHandler
	Firmware       *handler.FirmwareHandler
	Maintenance    *handler.MaintenanceHandler
	RobotConfig    *handler.RobotConfigHandler
	AlarmRule      *handler.AlarmRuleHandler
	Notify         *handler.NotifyHandler
	Camera         *handler.CameraHandler
	Prediction     *handler.PredictionHandler
	ReportTemplate *handler.ReportTemplateHandler
	Backup         *handler.BackupHandler
}

func Setup(hub *ws.Hub, publisher *mqtt.Publisher, taskScheduler *scheduler.TaskScheduler) *gin.Engine {
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS())

	// WebSocket endpoint
	r.GET("/ws", ws.HandleWebSocket(hub))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Handler instances
	h := &Handlers{
		User:           handler.NewUserHandler(),
		Station:        handler.NewStationHandler(),
		Robot:          handler.NewRobotHandler(publisher),
		Task:           handler.NewTaskHandler(publisher, taskScheduler),
		Alarm:          handler.NewAlarmHandler(),
		Monitor:        handler.NewMonitorHandler(),
		Analytics:      handler.NewAnalyticsHandler(),
		SystemConfig:   handler.NewSystemConfigHandler(),
		Dict:           handler.NewDictHandler(),
		Organization:   handler.NewOrganizationHandler(),
		Audit:          handler.NewAuditHandler(),
		Firmware:       handler.NewFirmwareHandler(publisher),
		Maintenance:    handler.NewMaintenanceHandler(),
		RobotConfig:    handler.NewRobotConfigHandler(),
		AlarmRule:      handler.NewAlarmRuleHandler(),
		Notify:         handler.NewNotifyHandler(),
		Camera:         handler.NewCameraHandler(),
		Prediction:     handler.NewPredictionHandler(),
		ReportTemplate: handler.NewReportTemplateHandler(),
		Backup:         handler.NewBackupHandler(),
	}

	v1 := r.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.User.Login)
		auth.GET("/captcha", h.User.GetCaptcha)
		auth.POST("/verify-captcha", h.User.VerifyCaptcha)
		auth.POST("/sms-code", h.User.SendSmsCode)
		auth.POST("/verify-sms", h.User.VerifySms)
		auth.POST("/forgot-password", h.User.ForgotPassword)
		auth.POST("/reset-password", h.User.ResetPassword)
	}

	// Protected routes (JWT + Audit)
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(), middleware.Audit())

	registerAuthRoutes(protected, h)
	registerStationRoutes(protected, h)
	registerRobotRoutes(protected, h)
	registerTaskRoutes(protected, h)
	registerAlarmRoutes(protected, h)
	registerMonitorRoutes(protected, h)
	registerAnalyticsRoutes(protected, h)
	registerUserRoutes(protected, h)
	registerSystemRoutes(protected, h)
	registerOrganizationRoutes(protected, h)
	registerDeviceRoutes(protected, h)
	registerReportRoutes(protected, h)

	return r
}
