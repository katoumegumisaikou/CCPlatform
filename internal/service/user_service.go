package service

import (
	"ccplatform/internal/config"
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"fmt"
	"time"
)

// UserService 用户与角色业务逻辑层，处理用户认证、RBAC 权限和系统初始化。
//
type UserService struct {
	repo     *repository.UserRepo
	roleRepo *repository.RoleRepo
}

// NewUserService 创建 UserService 实例。
func NewUserService() *UserService {
	return &UserService{
		repo:     repository.NewUserRepo(),
		roleRepo: repository.NewRoleRepo(),
	}
}

// Create 创建新用户，密码通过 bcrypt 加密存储。
func (s *UserService) Create(user *model.User, password string) error {
	hash, err := util.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = hash
	return s.repo.Create(user)
}

// GetByID 根据用户 ID 查询用户详情。
func (s *UserService) GetByID(id string) (*model.User, error) {
	return s.repo.GetByID(id)
}

// Update 更新用户信息。
func (s *UserService) Update(user *model.User) error {
	return s.repo.Update(user)
}

// Delete 删除用户。
func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}

// List 分页查询用户列表，支持按用户名或真实姓名搜索。
func (s *UserService) List(page, size int, keyword string) ([]model.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, keyword)
}

// Login 用户登录认证。
// 流程: 查找用户 → 校验状态 → 验证密码 → 更新登录时间 → 记录登录日志 → 生成 JWT Token。
func (s *UserService) Login(username, password, ip string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		s.recordLoginLog("", username, ip, 0, "user not found")
		return "", fmt.Errorf("user not found")
	}
	if user.Status == 0 {
		s.recordLoginLog(user.UserID, username, ip, 0, "user is disabled")
		return "", fmt.Errorf("user is disabled")
	}
	if !util.CheckPassword(password, user.PasswordHash) {
		s.recordLoginLog(user.UserID, username, ip, 0, "invalid password")
		return "", fmt.Errorf("invalid password")
	}
	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	s.repo.Update(user)

	// 记录登录成功日志
	s.recordLoginLog(user.UserID, username, ip, 1, "")

	// 生成 JWT Token，包含用户 ID、用户名和角色 ID
	token, err := util.GenerateToken(
		config.Cfg.JWT.Secret,
		user.UserID,
		user.Username,
		user.RoleID,
		config.Cfg.JWT.ExpireHours,
	)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

func (s *UserService) recordLoginLog(userID, username, ip string, result int8, failReason string) {
	logRepo := repository.NewLoginLogRepo()
	entry := &model.LoginLog{
		UserID:     userID,
		Username:   username,
		LoginIP:    ip,
		LoginTime:  time.Now(),
		LoginResult: result,
		FailReason: failReason,
	}
	go func() {
		_ = logRepo.Create(entry)
	}()
}

// GetUserFromToken 根据 Token 中的用户 ID 查询用户信息。
func (s *UserService) GetUserFromToken(userID string) (*model.User, error) {
	return s.repo.GetByID(userID)
}

// ===== 角色管理 =====

// CreateRole 创建新角色。
func (s *UserService) CreateRole(role *model.Role) error {
	return s.roleRepo.Create(role)
}

// GetRoles 查询所有角色列表。
func (s *UserService) GetRoles() ([]model.Role, error) {
	return s.roleRepo.List()
}

// GetRoleByID 根据角色 ID 查询角色详情。
func (s *UserService) GetRoleByID(id string) (*model.Role, error) {
	return s.roleRepo.GetByID(id)
}

// UpdateRole 更新角色信息。
func (s *UserService) UpdateRole(role *model.Role) error {
	_, err := s.roleRepo.GetByID(role.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	return s.roleRepo.Update(role)
}

// DeleteRole 删除角色。
func (s *UserService) DeleteRole(id string) error {
	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	return s.roleRepo.Delete(id)
}

// InitDefaultRoles 初始化系统预定义角色（5 个）。
// 仅在角色不存在时创建，已存在则跳过。系统启动时自动调用。
func (s *UserService) InitDefaultRoles() error {
	defaults := []model.Role{
		{RoleID: model.RoleSysAdmin, RoleName: "系统管理员", Description: "拥有系统全部权限", Permissions: `["*"]`},
		{RoleID: model.RoleStationMgr, RoleName: "电站管理员", Description: "负责电站管理", Permissions: `["station:*","robot:*","task:*","alarm:*","monitor:view","analytics:view"]`},
		{RoleID: model.RoleEngineer, RoleName: "运维工程师", Description: "设备维护和故障处理", Permissions: `["robot:view","robot:control","task:view","alarm:view","alarm:handle","monitor:view"]`},
		{RoleID: model.RoleOperator, RoleName: "监控值班员", Description: "日常监控和告警处理", Permissions: `["monitor:view","alarm:view","alarm:handle","task:view"]`},
		{RoleID: model.RoleViewer, RoleName: "普通用户", Description: "仅查看权限", Permissions: `["monitor:view","analytics:view"]`},
	}
	for _, role := range defaults {
		existing, _ := s.roleRepo.GetByID(role.RoleID)
		if existing == nil || existing.RoleID == "" {
			if err := s.roleRepo.Create(&role); err != nil {
				return err
			}
		}
	}
	return nil
}

// ForgotPassword 发起密码重置请求，生成重置 Token。
func (s *UserService) ForgotPassword(username string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}
	resetRepo := repository.NewPasswordResetRepo()
	reset := &model.PasswordReset{
		UserID:    user.UserID,
		Token:     util.GenerateID(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := resetRepo.Create(reset); err != nil {
		return "", err
	}
	return reset.Token, nil
}

// ResetPassword 使用重置 Token 修改密码。
func (s *UserService) ResetPassword(token, newPassword string) error {
	resetRepo := repository.NewPasswordResetRepo()
	pr, err := resetRepo.GetByToken(token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}
	if pr.Used == 1 {
		return fmt.Errorf("token already used")
	}
	if time.Now().After(pr.ExpiresAt) {
		return fmt.Errorf("token expired")
	}
	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(pr.UserID, hash); err != nil {
		return err
	}
	return resetRepo.MarkUsed(token)
}

// InitAdminUser 初始化默认管理员账号（admin/admin123）。
// 仅在 admin 用户不存在时创建，系统启动时自动调用。
func (s *UserService) InitAdminUser() error {
	_, err := s.repo.GetByUsername("admin")
	if err == nil {
		return nil // 已存在则跳过
	}
	admin := &model.User{
		UserID:   "admin01",
		Username: "admin",
		RealName: "系统管理员",
		RoleID:   model.RoleSysAdmin,
		Status:   1,
	}
	return s.Create(admin, "admin123")
}
