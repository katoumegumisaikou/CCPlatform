package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerReportRoutes(rg *gin.RouterGroup, h *Handlers) {
	reportTemplates := rg.Group("/report-templates")
	{
		reportTemplates.GET("", h.ReportTemplate.List)
		reportTemplates.POST("", middleware.RBAC("system:config"), h.ReportTemplate.Create)
		reportTemplates.PUT("/:id", middleware.RBAC("system:config"), h.ReportTemplate.Update)
		reportTemplates.DELETE("/:id", middleware.RBAC("system:config"), h.ReportTemplate.Delete)
	}

	backups := rg.Group("/backups")
	{
		backups.GET("", middleware.RBAC("system:config"), h.Backup.List)
		backups.POST("", middleware.RBAC("system:config"), h.Backup.Create)
		backups.DELETE("/:filename", middleware.RBAC("system:config"), h.Backup.Delete)
		backups.POST("/:filename/restore", middleware.RBAC("system:config"), h.Backup.Restore)
	}
}
