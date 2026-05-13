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
	ErrTaskRunning     = New(10010, "任务执行中", http.StatusBadRequest)
)
