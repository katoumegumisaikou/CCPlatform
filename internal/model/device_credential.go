package model

import "time"

// DeviceCredential MQTT 设备认证凭据模型。
type DeviceCredential struct {
	CredentialID  string    `gorm:"primaryKey;type:varchar(32)" json:"credential_id"`
	DeviceID      string    `gorm:"type:varchar(32);uniqueIndex" json:"device_id"`
	DeviceType    string    `gorm:"type:varchar(20)" json:"device_type"` // robot/camera/sensor
	Username      string    `gorm:"type:varchar(50)" json:"username"`
	PasswordHash  string    `gorm:"type:varchar(100)" json:"password_hash"`
	Token         string    `gorm:"type:varchar(200)" json:"token"`
	Status        int8      `gorm:"type:tinyint;default:1" json:"status"`
	LastConnected *time.Time `gorm:"type:datetime" json:"last_connected"`
	CreateTime    time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime    time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

func (DeviceCredential) TableName() string {
	return "device_credentials"
}

// UpgradeRecord OTA 升级记录模型。
type UpgradeRecord struct {
	RecordID    uint       `gorm:"primaryKey;autoIncrement" json:"record_id"`
	FirmwareID  string     `gorm:"type:varchar(32);index" json:"firmware_id"`
	RobotID     string     `gorm:"type:varchar(32);index" json:"robot_id"`
	FromVersion string     `gorm:"type:varchar(50)" json:"from_version"`
	ToVersion   string     `gorm:"type:varchar(50)" json:"to_version"`
	Status      int8       `gorm:"type:tinyint;default:0" json:"status"` // 0下发 1下载中 2安装中 3成功 4失败
	ErrorMsg    string     `gorm:"type:varchar(500)" json:"error_msg"`
	DispatchTime *time.Time `gorm:"type:datetime" json:"dispatch_time"`
	CompleteTime *time.Time `gorm:"type:datetime" json:"complete_time"`
	CreateTime  time.Time  `gorm:"autoCreateTime" json:"create_time"`
}

func (UpgradeRecord) TableName() string {
	return "upgrade_records"
}

// PasswordReset 密码重置模型。
type PasswordReset struct {
	ResetID    string    `gorm:"primaryKey;type:varchar(32)" json:"reset_id"`
	UserID     string    `gorm:"type:varchar(32);index" json:"user_id"`
	Token      string    `gorm:"type:varchar(200);uniqueIndex" json:"token"`
	ExpiresAt  time.Time `gorm:"type:datetime" json:"expires_at"`
	Used       int8      `gorm:"type:tinyint;default:0" json:"used"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}
