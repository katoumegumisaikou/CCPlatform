package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAlarmRoutes(rg *gin.RouterGroup, h *Handlers) {
	alarms := rg.Group("/alarms")
	{
		alarms.GET("", h.Alarm.List)
		alarms.GET("/stats", h.Alarm.GetStats)
		alarms.GET("/recent", h.Alarm.GetRecent)
		alarms.GET("/:id", h.Alarm.GetByID)
		alarms.PUT("/:id", middleware.RBAC("alarm:handle"), h.Alarm.Handle)
		alarms.GET("/trends", h.Alarm.GetTrendAnalysis)
	}

	alarmRules := rg.Group("/alarm-rules")
	{
		alarmRules.GET("", h.AlarmRule.List)
		alarmRules.POST("", middleware.RBAC("alarm:config"), h.AlarmRule.Create)
		alarmRules.PUT("/:id", middleware.RBAC("alarm:config"), h.AlarmRule.Update)
		alarmRules.DELETE("/:id", middleware.RBAC("alarm:config"), h.AlarmRule.Delete)
	}

	notify := rg.Group("/notify/templates")
	{
		notify.GET("", h.Notify.List)
		notify.POST("", middleware.RBAC("notify:config"), h.Notify.Create)
		notify.PUT("/:id", middleware.RBAC("notify:config"), h.Notify.Update)
		notify.DELETE("/:id", middleware.RBAC("notify:config"), h.Notify.Delete)
	}
}
