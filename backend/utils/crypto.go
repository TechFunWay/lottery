package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const passwordSalt = "caipiao"

// MD5Hash 将字符串进行MD5加密
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// MD5HashWithSalt 带盐值的MD5加密
func MD5HashWithSalt(text, salt string) string {
	combined := fmt.Sprintf("%s%s", text, salt)
	return MD5Hash(combined)
}

// HashPassword 密码加密：后端对前端传来的 MD5 值再做一次加盐 MD5
// 前端：md5(rawPassword)  ->  后端：md5(frontendMD5 + "caipiao")
func HashPassword(frontendMD5 string) string {
	return MD5HashWithSalt(frontendMD5, passwordSalt)
}
