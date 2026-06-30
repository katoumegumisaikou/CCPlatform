package model

import "time"

// LoginLog 登录日志模型，对应 login_logs 表。
// 记录每次登录尝试的结果，用于安全审计和异常检测。
type LoginLog struct {
	LogID      string    `gorm:"primaryKey;type:varchar(32)" json:"log_id"`
	UserID     string    `gorm:"type:varchar(32);index" json:"user_id"`
	Username   string    `gorm:"type:varchar(50)" json:"username"`
	LoginIP    string    `gorm:"type:varchar(50)" json:"login_ip"`
	LoginTime  time.Time `gorm:"type:datetime" json:"login_time"`
	LoginResult int8     `gorm:"type:tinyint" json:"login_result"` // 登录结果: 1成功 0失败
	FailReason string    `gorm:"type:varchar(200)" json:"fail_reason"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
