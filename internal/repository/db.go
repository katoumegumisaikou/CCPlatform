// Package repository 提供数据访问层，封装所有数据库操作。
// 每个 Repo 结构体对应一个数据模型，提供标准 CRUD 和业务查询方法。
// 所有 Repo 实例共享全局 DB 连接池，由 InitDB 在启动时初始化。
package repository

import (
	"ccplatform/internal/config"
	"ccplatform/internal/model"
	"fmt"

	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接实例，由 InitDB 初始化后供所有 Repo 使用。
var DB *gorm.DB

// InitDB 初始化 MySQL 数据库连接，配置连接池，并自动迁移所有数据模型。
// 该函数在 main.go 启动时调用，必须在任何 Repo 操作之前完成。
func InitDB() error {
	var err error
	DB, err = gorm.Open(mysql.Open(config.Cfg.Database.DSN()), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info), // 启用 SQL 日志，便于调试
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用迁移时的外键约束，避免与已有表结构冲突
	})
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns) // 最大空闲连接数
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns) // 最大打开连接数

	// GORM AutoMigrate 根据 struct 定义自动创建/更新表结构（不会删除已有列）
	if err := DB.AutoMigrate(
		&model.Station{},
		&model.Robot{},
		&model.Task{},
		&model.Alarm{},
		&model.User{},
		&model.Role{},
		&model.Firmware{},
		&model.Maintenance{},
		&model.AlarmRule{},
		&model.NotifyTemplate{},
		&model.AuditLog{},
		&model.LoginLog{},
		&model.SystemConfig{},
		&model.Dict{},
		&model.DictItem{},
		&model.Organization{},
		&model.RobotConfig{},
		&model.RobotPosition{},
		&model.EnvironmentData{},
		&model.Camera{},
		&model.CleaningRecord{},
		&model.ReportTemplate{},
		&model.DeviceCredential{},
		&model.UpgradeRecord{},
		&model.PasswordReset{},
		&model.Geofence{},
		&model.GeofenceAlarm{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接池，应在优雅关闭阶段调用。
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
