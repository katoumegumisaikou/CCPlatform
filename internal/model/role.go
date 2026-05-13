package model

import "time"

// Role 角色模型，对应 roles 表。
// 采用 RBAC (基于角色的访问控制) 模型，Permissions 字段存储 JSON 格式的权限列表。
type Role struct {
	RoleID      string    `gorm:"primaryKey;type:varchar(32)" json:"role_id"`     // 角色唯一标识
	RoleName    string    `gorm:"type:varchar(50);uniqueIndex" json:"role_name"`  // 角色名称（唯一）
	Description string    `gorm:"type:varchar(200)" json:"description"`          // 角色描述
	Permissions string    `gorm:"type:text" json:"permissions"`                  // 权限列表（JSON 数组，如 ["station:*","robot:view"]）
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`             // 创建时间
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time"`             // 更新时间
}

func (Role) TableName() string {
	return "roles"
}

// 预定义角色常量，系统启动时自动创建。
// 权限格式: "模块:操作"，如 "station:create"，"*" 表示全部权限。
const (
	RoleSysAdmin   = "sys_admin"   // 系统管理员 - 拥有全部权限
	RoleStationMgr = "station_mgr" // 电站管理员 - 管理所属电站的所有资源
	RoleEngineer   = "engineer"    // 运维工程师 - 设备维护和故障处理
	RoleOperator   = "operator"    // 监控值班员 - 日常监控和告警处理
	RoleViewer     = "viewer"      // 普通用户 - 仅查看权限
)
