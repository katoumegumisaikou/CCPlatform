package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(rg *gin.RouterGroup, h *Handlers) {
	users := rg.Group("/users")
	{
		users.GET("", middleware.RBAC("user:view"), h.User.List)
		users.POST("", middleware.RBAC("user:create"), h.User.Create)
		users.PUT("/:id", middleware.RBAC("user:update"), h.User.Update)
		users.DELETE("/:id", middleware.RBAC("user:delete"), h.User.Delete)
	}

	roles := rg.Group("/roles")
	{
		roles.GET("", h.User.ListRoles)
		roles.POST("", middleware.RBAC("role:create"), h.User.CreateRole)
		roles.PUT("/:id", middleware.RBAC("role:update"), h.User.UpdateRole)
		roles.DELETE("/:id", middleware.RBAC("role:delete"), h.User.DeleteRole)
	}

	loginLogs := rg.Group("/login-logs")
	{
		loginLogs.GET("", h.User.ListLoginLogs)
	}
}
