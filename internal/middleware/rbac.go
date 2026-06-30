package middleware

// RBAC 基于角色的访问控制中间件，sys_admin 拥有全部权限，其他角色按模块权限矩阵校验。

import (
	"ccplatform/internal/model"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"

	"github.com/gin-gonic/gin"
)

// RBAC 权限检查中间件
func RBAC(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, exists := c.Get("role_id")
		if !exists {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}

		role := roleID.(string)

		// 系统管理员拥有所有权限
		if role == model.RoleSysAdmin {
			c.Next()
			return
		}

		// 检查角色权限 (简化版，实际应从数据库读取并解析JSON)
		if !checkPermission(role, requiredPermission) {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkPermission(roleID, permission string) bool {
	// 简化的权限矩阵
	permissions := map[string][]string{
		model.RoleStationMgr: {"station", "robot", "task", "alarm", "monitor", "analytics"},
		model.RoleEngineer:   {"robot", "task", "alarm", "monitor"},
		model.RoleOperator:   {"monitor", "alarm", "task"},
		model.RoleViewer:     {"monitor", "analytics"},
	}

	allowed, ok := permissions[roleID]
	if !ok {
		return false
	}

	// 检查模块权限
	module := permission
	if idx := len(permission); idx > 0 {
		for i, c := range permission {
			if c == ':' {
				module = permission[:i]
				break
			}
		}
	}

	for _, p := range allowed {
		if p == module || p == "*" {
			return true
		}
	}
	return false
}
