package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// RandToken 返回 n 字节加密随机数的十六进制串，用于生成实体 ID。
func RandToken(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand 失败属极端环境事件，回退到时间戳哈希仍保证唯一性语义。
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(buf)
}
