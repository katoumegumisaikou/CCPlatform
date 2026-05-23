package middleware

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"strings"

	"github.com/gin-gonic/gin"
)

// Audit 操作审计中间件，记录非 GET 请求的操作日志。
func Audit() gin.HandlerFunc {
	auditRepo := repository.NewAuditLogRepo()

	return func(c *gin.Context) {
		c.Next()

		// 只记录写操作（POST/PUT/DELETE）
		method := c.Request.Method
		if method == "GET" {
			return
		}

		operatorID := c.GetString("user_id")
		operatorName := c.GetString("username")

		var operationType string
		switch method {
		case "POST":
			operationType = "create"
		case "PUT":
			operationType = "update"
		case "DELETE":
			operationType = "delete"
		default:
			operationType = strings.ToLower(method)
		}

		// 简化的操作描述：METHOD /path
		targetResource := method + " " + c.Request.URL.Path

		logEntry := &model.AuditLog{
			LogID:          util.GenerateID(),
			OperatorID:     operatorID,
			OperatorName:   operatorName,
			OperationType:  operationType,
			TargetResource: targetResource,
			IP:             c.ClientIP(),
		}

		// 异步记录，不阻塞请求响应
		go func() {
			_ = auditRepo.Create(logEntry)
		}()
	}
}
