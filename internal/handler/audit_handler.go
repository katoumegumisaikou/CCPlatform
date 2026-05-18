package handler

import (
	"ccplatform/internal/repository"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AuditHandler 操作日志 HTTP 处理器。
type AuditHandler struct {
	repo *repository.AuditLogRepo
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{repo: repository.NewAuditLogRepo()}
}

// List 分页查询操作日志。
// GET /api/v1/audit-logs?page=1&size=10&operator_id=xx&operation_type=xx&start_time=xx&end_time=xx
func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	operatorID := c.Query("operator_id")
	operationType := c.Query("operation_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	logs, total, err := h.repo.List(page, size, operatorID, operationType, startTime, endTime)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, logs, total, page, size)
}
