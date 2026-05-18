package handler

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/internal/service"
	"ccplatform/pkg/errcode"
	"ccplatform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户与角色 HTTP 处理器，处理认证、用户管理和角色管理。
//
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
	token, err := h.svc.Login(req.Username, req.Password, c.ClientIP())
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

// UpdateRole 更新角色接口（部分更新）。
// PUT /api/v1/roles/:id
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.svc.GetRoleByID(id)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if req.RoleName != "" {
		existing.RoleName = req.RoleName
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Permissions != "" {
		existing.Permissions = req.Permissions
	}
	existing.RoleID = id
	if err := h.svc.UpdateRole(existing); err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, existing)
}

// DeleteRole 删除角色接口。
// DELETE /api/v1/roles/:id
func (h *UserHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteRole(id); err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.OK(c, nil)
}

// ListLoginLogs 登录日志列表接口。
// GET /api/v1/login-logs?page=1&size=10&user_id=xx&start_time=xx&end_time=xx
func (h *UserHandler) ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	userID := c.Query("user_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	loginLogRepo := repository.NewLoginLogRepo()
	logs, total, err := loginLogRepo.List(page, size, userID, startTime, endTime)
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OKPage(c, logs, total, page, size)
}

// ForgotPasswordRequest 忘记密码请求体。
type ForgotPasswordRequest struct {
	Username string `json:"username" binding:"required"`
}

// ForgotPassword 忘记密码接口。
// POST /api/v1/auth/forgot-password
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	token, err := h.svc.ForgotPassword(req.Username)
	if err != nil {
		response.ErrorMsg(c, 400, 10001, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "reset token generated", "reset_token": token})
}

// ResetPasswordRequest 重置密码请求体。
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword 重置密码接口。
// POST /api/v1/auth/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if err := h.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.ErrorMsg(c, 400, 10001, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "password reset successfully"})
}

// CaptchaService is a global instance for the captcha service.
var captchaSvc = service.NewCaptchaService()

// GetCaptcha 获取图形验证码接口。
// GET /api/v1/auth/captcha
func (h *UserHandler) GetCaptcha(c *gin.Context) {
	id, question, _, err := captchaSvc.GenerateCaptcha()
	if err != nil {
		response.Error(c, errcode.ErrInternal)
		return
	}
	response.OK(c, gin.H{"captcha_id": id, "question": question})
}

// VerifyCaptchaRequest 验证码校验请求体。
type VerifyCaptchaRequest struct {
	CaptchaID string `json:"captcha_id" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

// VerifyCaptcha 验证图形验证码接口。
// POST /api/v1/auth/verify-captcha
func (h *UserHandler) VerifyCaptcha(c *gin.Context) {
	var req VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if !captchaSvc.VerifyCaptcha(req.CaptchaID, req.Answer) {
		response.Error(c, errcode.ErrCaptchaError)
		return
	}
	response.OK(c, gin.H{"message": "verified"})
}

// SmsCodeRequest 短信验证码请求体。
type SmsCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// SendSmsCode 发送短信验证码接口。
// POST /api/v1/auth/sms-code
func (h *UserHandler) SendSmsCode(c *gin.Context) {
	var req SmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	code, err := captchaSvc.GenerateSMSCode(req.Phone)
	if err != nil {
		response.Error(c, errcode.ErrSmsSendFailed)
		return
	}
	response.OK(c, gin.H{"message": "sms code sent", "code": code})
}

// VerifySmsRequest 短信验证码校验请求体。
type VerifySmsRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// VerifySms 验证短信验证码接口。
// POST /api/v1/auth/verify-sms
func (h *UserHandler) VerifySms(c *gin.Context) {
	var req VerifySmsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrParam)
		return
	}
	if !captchaSvc.VerifySMS(req.Phone, req.Code) {
		response.Error(c, errcode.ErrCaptchaError)
		return
	}
	response.OK(c, gin.H{"message": "verified"})
}
