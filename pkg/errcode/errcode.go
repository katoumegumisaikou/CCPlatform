// Package errcode 定义全局错误码，统一 API 错误响应格式。
package errcode

import "net/http"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, msg string, status int) *AppError {
	return &AppError{Code: code, Message: msg, Status: status}
}

var (
	Success            = New(0, "success", http.StatusOK)
	ErrParam           = New(10001, "参数错误", http.StatusBadRequest)
	ErrUnauthorized    = New(10002, "未授权", http.StatusUnauthorized)
	ErrForbidden       = New(10003, "权限不足", http.StatusForbidden)
	ErrNotFound        = New(10004, "资源不存在", http.StatusNotFound)
	ErrInternal        = New(10005, "内部错误", http.StatusInternalServerError)
	ErrDuplicate       = New(10006, "数据已存在", http.StatusConflict)
	ErrLoginFailed     = New(10007, "用户名或密码错误", http.StatusUnauthorized)
	ErrTokenExpired    = New(10008, "Token已过期", http.StatusUnauthorized)
	ErrRobotOffline    = New(10009, "机器人离线", http.StatusBadRequest)
	ErrTaskRunning         = New(10010, "任务执行中", http.StatusBadRequest)
	ErrCaptchaError        = New(10011, "验证码错误", http.StatusBadRequest)
	ErrSmsSendFailed       = New(10012, "短信发送失败", http.StatusInternalServerError)
	ErrOtaInProgress       = New(10013, "固件升级进行中", http.StatusBadRequest)
	ErrBackupFailed        = New(10014, "备份失败", http.StatusInternalServerError)
	ErrFirmwareNotFound    = New(10015, "固件不存在", http.StatusNotFound)
	ErrMaintenanceNotFound = New(10016, "维护记录不存在", http.StatusNotFound)
	ErrResetTokenExpired   = New(10017, "重置令牌已过期", http.StatusBadRequest)
	ErrNoAvailableRobot    = New(10018, "无可用机器人", http.StatusBadRequest)
	ErrOrgHasChildren      = New(10019, "组织下有子节点，无法删除", http.StatusBadRequest)
	ErrPredictionNotFound  = New(10022, "故障预测记录不存在", http.StatusNotFound)
)
