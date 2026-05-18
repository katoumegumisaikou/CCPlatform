// Package config 负责加载和管理平台配置。
//
// 配置文件使用 YAML 格式，通过 Viper 库读取并反序列化到 Config 结构体。
// 全局变量 Cfg 在 Load() 后可被其他包直接引用。
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 是平台的顶层配置结构，对应 config.yaml 的根节点。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`   // HTTP 服务器配置
	Database DatabaseConfig `mapstructure:"database"` // MySQL 数据库配置
	MQTT     MQTTConfig     `mapstructure:"mqtt"`     // MQTT Broker 配置
	JWT      JWTConfig      `mapstructure:"jwt"`      // JWT 认证配置
	Log      LogConfig      `mapstructure:"log"`      // 日志配置
	Notify   NotifyConfig   `mapstructure:"notify"`   // 通知渠道配置
	OTA      OTAConfig      `mapstructure:"ota"`      // OTA 固件升级配置
	Backup   BackupConfig   `mapstructure:"backup"`   // 数据备份配置
	Video    VideoConfig    `mapstructure:"video"`    // 视频监控配置
}

// ServerConfig HTTP 服务器配置。
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口，默认 8080
	Mode string `mapstructure:"mode"` // 运行模式: debug/release/test
}

// DatabaseConfig MySQL 数据库配置。
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`          // 数据库主机地址
	Port         int    `mapstructure:"port"`          // 数据库端口
	User         string `mapstructure:"user"`          // 用户名
	Password     string `mapstructure:"password"`      // 密码
	DBName       string `mapstructure:"dbname"`        // 数据库名
	Charset      string `mapstructure:"charset"`       // 字符集，默认 utf8mb4
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 最大打开连接数
}

// DSN 生成 MySQL 连接字符串，格式: user:password@tcp(host:port)/dbname?charset=xxx&parseTime=True&loc=Local
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

// MQTTConfig MQTT Broker 配置。
type MQTTConfig struct {
	Port   int `mapstructure:"port"`     // MQTT TCP 端口，默认 1883
	WSPort int `mapstructure:"ws_port"`  // MQTT WebSocket 端口，默认 8083
}

// JWTConfig JWT 认证配置。
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`       // Token 签名密钥
	ExpireHours int    `mapstructure:"expire_hours"` // Token 过期时间（小时）
}

// LogConfig 日志配置。
type LogConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别: debug/info/warn/error
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个日志文件最大大小 (MB)
	MaxBackups int    `mapstructure:"max_backups"` // 最大备份文件数
	MaxAge     int    `mapstructure:"max_age"`     // 日志保留天数
}

// NotifyConfig 通知渠道配置，包括短信/邮件/APP推送的 API 密钥和模板。
type NotifyConfig struct {
	SMS struct {
		Provider  string `mapstructure:"provider"`  // 短信服务商: aliyun/tencent
		AccessKey string `mapstructure:"access_key"` // API 访问密钥
		SecretKey string `mapstructure:"secret_key"` // API 密钥
		SignName  string `mapstructure:"sign_name"`  // 短信签名
	} `mapstructure:"sms"`
	Email struct {
		SMTPHost string `mapstructure:"smtp_host"` // SMTP 服务器地址
		SMTPPort int    `mapstructure:"smtp_port"` // SMTP 端口
		From     string `mapstructure:"from"`      // 发件人邮箱
		Password string `mapstructure:"password"`  // 邮箱密码/授权码
	} `mapstructure:"email"`
	AppPush struct {
		Provider string `mapstructure:"provider"` // 推送服务商: jpush/getui
		AppKey   string `mapstructure:"app_key"`
		Secret   string `mapstructure:"secret"`
	} `mapstructure:"app_push"`
}

// OTAConfig OTA 固件升级配置。
type OTAConfig struct {
	StoragePath    string `mapstructure:"storage_path"`    // 固件文件存储路径
	UpgradeTimeout int    `mapstructure:"upgrade_timeout"` // 升级超时时间（秒）
	MaxRetry       int    `mapstructure:"max_retry"`       // 最大重试次数
}

// BackupConfig 数据备份配置。
type BackupConfig struct {
	BackupPath    string `mapstructure:"backup_path"`    // 备份文件存储路径
	CronSchedule  string `mapstructure:"cron_schedule"`  // 自动备份 cron 表达式
	RetentionDays int    `mapstructure:"retention_days"` // 备份保留天数
}

// VideoConfig 视频监控配置。
type VideoConfig struct {
	StreamType string `mapstructure:"stream_type"` // 默认流转发类型: rtsp/hls/webrtc
	MaxViewers int    `mapstructure:"max_viewers"` // 每路视频最大同时观看数
}

// Cfg 是全局配置实例，Load() 成功后可直接使用 config.Cfg.XXX 访问。
var Cfg Config

// Load 从指定路径读取 YAML 配置文件并反序列化到全局 Cfg 变量。
func Load(path string) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}
