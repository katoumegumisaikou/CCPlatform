package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户与角色 HTTP 处理器，处理认证、用户管理和角色管理。
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建 UserHandler 实例。
func NewUserHandler() *UserHandler {
	return &UserHandler{svc: service.NewUserService()}
}

// LoginRequest 登录请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名（必填）
	Password string `json:"password" binding:"required"` // 密码（必填）
}

// Login 用户登录接口，验证凭据后返回 JWT Token。
// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, errcode.ErrLoginFailed)
		return
	}
	response.OK(c, gin.H{"token": token})
}

// GetCurrentUser 获取当前登录用户信息，从 JWT Token 中解析用户 ID。
// GET /api/v1/auth/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id") // 由 JWT 中间件注入
	user, err := h.svc.GetByID(userID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, user)
}

// List 用户列表接口，支持分页和关键词搜索。
// GET /api/v1/users?page=1&size=10&keyword=xxx
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	users, total, err := h.svc.List(page, size, keyword)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, users, total, page, size)
}

// CreateUserRequest 创建用户请求体。
type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"` // 用户名（必填）
	Password   string `json:"password" binding:"required"` // 密码（必填）
	RealName   string `json:"real_name"`                   // 真实姓名
	Phone      string `json:"phone"`                       // 手机号
	Email      string `json:"email"`                       // 邮箱
	RoleID     string `json:"role_id"`                     // 角色 ID
	StationIDs string `json:"station_ids"`                 // 授权电站 ID 列表
}

// Create 创建用户接口，密码通过 bcrypt 加密存储。
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	user := &model.User{
		Username:   req.Username,
		RealName:   req.RealName,
		Phone:      req.Phone,
		Email:      req.Email,
		RoleID:     req.RoleID,
		StationIDs: req.StationIDs,
		Status:     1,
	}
	if err := h.svc.Create(user, req.Password); err != nil {
		response.Error(c, errcode.ErrDuplicate)
		return
	}
	response.OK(c, user)
}

// UpdateUserRequest 更新用户请求体，所有字段可选。
type UpdateUserRequest struct {
	RealName   string `json:"real_name"`  // 真实姓名
	Phone      string `json:"phone"`      // 手机号
	Email      string `json:"email"`      // 邮箱
	RoleID     string `json:"role_id"`    // 角色 ID
	StationIDs string `json:"station_ids"` // 授权电站 ID 列表
	Status     *int8  `json:"status"`     // 状态(0禁用 1启用)
}

// Update 更新用户接口，采用部分更新策略。
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	user, err := h.svc.GetByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != "" {
		user.RoleID = req.RoleID
	}
	if req.StationIDs != "" {
		user.StationIDs = req.StationIDs
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if err := h.svc.Update(user); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, user)
}

// Delete 删除用户接口。
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, nil)
}

// ===== 角色管理接口 =====

// ListRoles 角色列表接口，返回所有角色及权限配置。
// GET /api/v1/roles
func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.GetRoles()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, roles)
}

// CreateRoleRequest 创建角色请求体。
type CreateRoleRequest struct {
	RoleName    string `json:"role_name" binding:"required"` // 角色名称（必填）
	Description string `json:"description"`                  // 角色描述
	Permissions string `json:"permissions"`                  // 权限列表(JSON 数组)
}

// CreateRole 创建角色接口。
// POST /api/v1/roles
func (h *UserHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	role := &model.Role{
		RoleName:    req.RoleName,
		Description: req.Description,
		Permissions: req.Permissions,
	}
	if err := h.svc.CreateRole(role); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, role)
}
