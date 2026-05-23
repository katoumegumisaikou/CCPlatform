package util

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateID 生成 32 字符的随机十六进制 ID（UUID v4 去除连字符）。
func GenerateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
