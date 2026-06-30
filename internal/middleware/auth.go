// Package middleware 提供 HTTP 中间件，包括 JWT 认证、RBAC 权限检查和 CORS 跨域。
package middleware

import (
	"ccplatform/internal/config"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"ccplatform/pkg/util"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件，解析 Authorization 头中的 Bearer Token，
// 验证通过后将 user_id/username/role_id 注入 Gin Context。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := util.ParseToken(config.Cfg.JWT.Secret, parts[1])
		if err != nil {
			response.Error(c, errcode.ErrTokenExpired)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_id", claims.RoleID)
		c.Next()
	}
}
