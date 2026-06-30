package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DictHandler 数据字典 HTTP 处理器。
type DictHandler struct {
	svc *service.DictService
}

func NewDictHandler() *DictHandler {
	return &DictHandler{svc: service.NewDictService()}
}

// --- Dict ---

// ListDict 分页查询字典类型列表。
// GET /api/v1/dicts?page=1&size=10
func (h *DictHandler) ListDict(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := h.svc.ListDict(page, size)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// CreateDict 创建字典类型。
// POST /api/v1/dicts
func (h *DictHandler) CreateDict(c *gin.Context) {
	var req model.Dict
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.CreateDict(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// UpdateDict 更新字典类型。
// PUT /api/v1/dicts/:id
func (h *DictHandler) UpdateDict(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.svc.GetDictByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req model.Dict
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.DictName != "" {
		existing.DictName = req.DictName
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	existing.DictID = id
	if err := h.svc.UpdateDict(existing); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, existing)
}

// DeleteDict 删除字典类型及其所有字典项。
// DELETE /api/v1/dicts/:id
func (h *DictHandler) DeleteDict(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteDict(id); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, nil)
}

// --- DictItem ---

// ListItems 查询指定字典类型下的所有字典项。
// GET /api/v1/dicts/:type/items
func (h *DictHandler) ListItems(c *gin.Context) {
	dictType := c.Param("type")
	items, err := h.svc.GetItemsByType(dictType)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, items)
}

// CreateItem 创建字典项。
// POST /api/v1/dicts/:type/items
func (h *DictHandler) CreateItem(c *gin.Context) {
	dictType := c.Param("type")
	var req model.DictItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	req.DictType = dictType
	if err := h.svc.CreateItem(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// UpdateItem 更新字典项。
// PUT /api/v1/dicts/items/:id
func (h *DictHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.svc.GetItemByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req model.DictItem
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.ItemKey != "" {
		existing.ItemKey = req.ItemKey
	}
	if req.ItemValue != "" {
		existing.ItemValue = req.ItemValue
	}
	existing.ItemID = id
	if err := h.svc.UpdateItem(existing); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, existing)
}

// DeleteItem 删除字典项。
// DELETE /api/v1/dicts/items/:id
func (h *DictHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteItem(id); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, nil)
}
