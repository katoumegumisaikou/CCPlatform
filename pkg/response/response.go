// Package response 提供统一的 JSON 响应格式，包含 code/message/data 三个字段。
package response

import (
	"ccplatform/pkg/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: errcode.Success.Message,
		Data:    data,
	})
}

func OKPage(c *gin.Context, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success.Code,
		Message: errcode.Success.Message,
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

func Error(c *gin.Context, err *errcode.AppError) {
	c.JSON(err.Status, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

func ErrorMsg(c *gin.Context, status int, code int, msg string) {
	c.JSON(status, Response{
		Code:    code,
		Message: msg,
	})
}
