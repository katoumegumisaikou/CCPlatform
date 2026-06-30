package handler

import (
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"

	"github.com/gin-gonic/gin"
)

// BackupHandler 数据备份 HTTP 处理器。
type BackupHandler struct {
	svc *service.BackupService
}

func NewBackupHandler() *BackupHandler {
	return &BackupHandler{svc: service.NewBackupService()}
}

// List 备份文件列表接口。
// GET /api/v1/backups
func (h *BackupHandler) List(c *gin.Context) {
	files, err := h.svc.List()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, files)
}

// Create 创建备份接口。
// POST /api/v1/backups
func (h *BackupHandler) Create(c *gin.Context) {
	file, err := h.svc.Create()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, file)
}

// Restore 恢复备份接口。
// POST /api/v1/backups/:filename/restore
func (h *BackupHandler) Restore(c *gin.Context) {
	filename := c.Param("filename")
	if err := h.svc.Restore(filename); err != nil {
		response.ErrorMsg(c, 400, 10014, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "restore started"})
}

// Delete 删除备份文件接口。
// DELETE /api/v1/backups/:filename
func (h *BackupHandler) Delete(c *gin.Context) {
	filename := c.Param("filename")
	if err := h.svc.Delete(filename); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}
