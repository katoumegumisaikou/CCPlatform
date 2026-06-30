package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户数据访问层，封装 users 表的所有数据库操作。
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建 UserRepo 实例，使用全局 DB 连接。
func NewUserRepo() *UserRepo {
	return &UserRepo{db: DB}
}

// Create 创建新用户记录。
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据用户 ID 查询单条记录。
func (r *UserRepo) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("user_id = ?", id).First(&user).Error
	return &user, err
}

// GetByUsername 根据用户名查询用户，用于登录认证。
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// Update 更新用户信息（全量更新）。
func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdatePassword 更新用户密码哈希。
func (r *UserRepo) UpdatePassword(userID, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("user_id = ?", userID).Update("password_hash", passwordHash).Error
}

// Delete 根据用户 ID 删除记录。
func (r *UserRepo) Delete(id string) error {
	return r.db.Where("user_id = ?", id).Delete(&model.User{}).Error
}

// List 分页查询用户列表，支持按用户名或真实姓名模糊搜索。
func (r *UserRepo) List(page, size int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := r.db.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Role").Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&users).Error
	return users, total, err
}

// ===== 角色数据访问层 =====

// RoleRepo 角色数据访问层，封装 roles 表的所有数据库操作。
type RoleRepo struct {
	db *gorm.DB
}

// NewRoleRepo 创建 RoleRepo 实例，使用全局 DB 连接。
func NewRoleRepo() *RoleRepo {
	return &RoleRepo{db: DB}
}

// Create 创建新角色记录。
func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// GetByID 根据角色 ID 查询单条记录。
func (r *RoleRepo) GetByID(id string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("role_id = ?", id).First(&role).Error
	return &role, err
}

// Update 更新角色信息（全量更新）。
func (r *RoleRepo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// List 查询所有角色列表，用于用户管理页面的角色下拉选择。
func (r *RoleRepo) List() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

// Delete 根据角色 ID 删除记录。
func (r *RoleRepo) Delete(id string) error {
	return r.db.Where("role_id = ?", id).Delete(&model.Role{}).Error
}
