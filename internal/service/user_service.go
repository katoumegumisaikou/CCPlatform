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
// TODO: 密码重置流程 — 忘记密码/邮件重置/短信验证码重置 (需求 7.1)
// TODO: 组织架构管理 — 多级组织结构(公司/区域/电站)，与电站关联 (需求 4.6.2)
// TODO: 系统配置/字典 — 平台参数 CRUD、数据字典管理 (需求 4.6.2)
// TODO: 操作日志审计 — 记录用户操作/登录日志，支持查询导出 (需求 4.6.2)
// TODO: 数据备份 — 自动/手动备份策略，数据恢复 (需求 4.6.2)
// TODO: 多因子认证 — 验证码/短信验证码登录 (需求 7.1)
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
// 流程: 查找用户 → 校验状态 → 验证密码 → 更新登录时间 → 生成 JWT Token。
func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}
	if user.Status == 0 {
		return "", fmt.Errorf("user is disabled")
	}
	if !util.CheckPassword(password, user.PasswordHash) {
		return "", fmt.Errorf("invalid password")
	}
	// 更新最后登录时间
	user.LastLogin = time.Now()
	s.repo.Update(user)

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
