package model

import "time"

// User 系统用户模型，对应 users 表。
// 采用 RBAC 权限模型，通过 RoleID 关联角色，StationIDs 控制可访问的电站范围。
// PasswordHash 使用 bcrypt 加密，JSON 序列化时隐藏（json:"-"）。
type User struct {
	UserID       string     `gorm:"primaryKey;type:varchar(32)" json:"user_id"`                // 用户唯一标识
	Username     string     `gorm:"type:varchar(50);uniqueIndex" json:"username"`              // 用户名（唯一）
	PasswordHash string     `gorm:"type:varchar(100);not null" json:"-"`                       // 密码哈希（bcrypt），API 不返回
	RealName     string     `gorm:"type:varchar(50)" json:"real_name"`                         // 真实姓名
	Phone        string     `gorm:"type:varchar(20)" json:"phone"`                             // 手机号
	Email        string     `gorm:"type:varchar(100)" json:"email"`                            // 邮箱
	RoleID       string     `gorm:"type:varchar(32);index" json:"role_id"`                     // 角色 ID（关联 roles 表）
	Role         *Role      `gorm:"foreignKey:RoleID;references:RoleID" json:"role,omitempty"` // 角色信息
	StationIDs   string     `gorm:"type:varchar(500)" json:"station_ids"`                      // 授权电站 ID 列表（逗号分隔）
	Status       int8       `gorm:"type:tinyint;default:1" json:"status"`                      // 状态: 0禁用 1启用
	LastLogin    *time.Time `gorm:"type:datetime" json:"last_login"`                           // 最后登录时间
	CreateTime   time.Time  `gorm:"autoCreateTime" json:"create_time"`                         // 创建时间
	UpdateTime   time.Time  `gorm:"autoUpdateTime" json:"update_time"`                         // 更新时间
}

func (User) TableName() string {
	return "users"
}
