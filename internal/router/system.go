package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(rg *gin.RouterGroup, h *Handlers) {
	systemConfigs := rg.Group("/system/configs")
	{
		systemConfigs.GET("", h.SystemConfig.List)
		systemConfigs.GET("/:key", h.SystemConfig.Get)
		systemConfigs.PUT("/:key", middleware.RBAC("system:config"), h.SystemConfig.Set)
		systemConfigs.DELETE("/:key", middleware.RBAC("system:config"), h.SystemConfig.Delete)
	}

	dicts := rg.Group("/dicts")
	{
		dicts.GET("", h.Dict.ListDict)
		dicts.POST("", middleware.RBAC("system:config"), h.Dict.CreateDict)
		dicts.PUT("/:id", middleware.RBAC("system:config"), h.Dict.UpdateDict)
		dicts.DELETE("/:id", middleware.RBAC("system:config"), h.Dict.DeleteDict)
		dicts.GET("/:type/items", h.Dict.ListItems)
		dicts.POST("/:type/items", middleware.RBAC("system:config"), h.Dict.CreateItem)
		dicts.PUT("/items/:id", middleware.RBAC("system:config"), h.Dict.UpdateItem)
		dicts.DELETE("/items/:id", middleware.RBAC("system:config"), h.Dict.DeleteItem)
	}

	auditLogs := rg.Group("/audit-logs")
	{
		auditLogs.GET("", h.Audit.List)
	}
}
