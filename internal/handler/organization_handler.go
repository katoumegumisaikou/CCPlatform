package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// OrganizationHandler 组织架构 HTTP 处理器。
type OrganizationHandler struct {
	svc *service.OrganizationService
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{svc: service.NewOrganizationService()}
}

// List 分页查询组织列表。
// GET /api/v1/organizations?page=1&size=10&org_type=xx&keyword=xx
func (h *OrganizationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	orgType := c.Query("org_type")
	keyword := c.Query("keyword")

	list, total, err := h.svc.List(page, size, orgType, keyword)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, list, total, page, size)
}

// GetTree 查询组织架构树。
// GET /api/v1/organizations/tree
func (h *OrganizationHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetTree()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, tree)
}

// Create 创建组织。
// POST /api/v1/organizations
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req model.Organization
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, req)
}

// GetByID 查询单个组织。
// GET /api/v1/organizations/:id
func (h *OrganizationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	org, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, org)
}

// Update 更新组织信息。
// PUT /api/v1/organizations/:id
func (h *OrganizationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req model.Organization
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.OrgName != "" {
		existing.OrgName = req.OrgName
	}
	if req.OrgCode != "" {
		existing.OrgCode = req.OrgCode
	}
	if req.OrgType != "" {
		existing.OrgType = req.OrgType
	}
	if req.ParentID != "" {
		existing.ParentID = req.ParentID
	}
	if req.StationIDs != "" {
		existing.StationIDs = req.StationIDs
	}
	existing.OrgID = id
	if err := h.svc.Update(existing); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, existing)
}

// Delete 删除组织（有子节点不可删除）。
// DELETE /api/v1/organizations/:id
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.ErrorMsg(c, 400, 10019, err.Error())
		return
	}
	response.OK(c, nil)
}
