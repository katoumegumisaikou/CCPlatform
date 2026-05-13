// Package repository 提供数据访问层，封装所有数据库操作。
// 每个 Repo 结构体对应一个数据模型，提供标准 CRUD 和业务查询方法。
// 所有 Repo 实例共享全局 DB 连接池，由 InitDB 在启动时初始化。
package repository

import (
	"ccplatform/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 是全局数据库连接实例，由 InitDB 初始化后供所有 Repo 使用。
var DB *gorm.DB

// InitDB 初始化 MySQL 数据库连接。
// 从配置文件读取 DSN，创建 GORM 连接池，并设置最大空闲/活跃连接数。
// 该函数在 main.go 启动时调用，必须在任何 Repo 操作之前完成。
func InitDB() error {
	var err error
	DB, err = gorm.Open(mysql.Open(config.Cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 启用 SQL 日志，便于调试
	})
	if err != nil {
		return err
	}
	// 获取底层 sql.DB 以配置连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns) // 最大空闲连接数
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns) // 最大打开连接数
	return nil
}
