package router

import (
	"ccplatform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerOrganizationRoutes(rg *gin.RouterGroup, h *Handlers) {
	orgs := rg.Group("/organizations")
	{
		orgs.GET("", h.Organization.List)
		orgs.GET("/tree", h.Organization.GetTree)
		orgs.GET("/:id", h.Organization.GetByID)
		orgs.POST("", middleware.RBAC("org:create"), h.Organization.Create)
		orgs.PUT("/:id", middleware.RBAC("org:update"), h.Organization.Update)
		orgs.DELETE("/:id", middleware.RBAC("org:delete"), h.Organization.Delete)
	}
}
