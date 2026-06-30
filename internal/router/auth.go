package router

import "github.com/gin-gonic/gin"

func registerAuthRoutes(rg *gin.RouterGroup, h *Handlers) {
	rg.GET("/auth/me", h.User.GetCurrentUser)
}
