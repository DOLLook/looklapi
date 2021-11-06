package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// 判断字符串是否为空
func IsEmpty(str string) bool {
	str = strings.TrimSpace(str)
	return len(str) == 0
}

// 获取字符MD5值
func MD5(str string) string {
	_md5 := md5.New()
	_md5.Write([]byte(str))
	return hex.EncodeToString(_md5.Sum([]byte(nil)))
}

// base64编码
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// base64解码
func Base64Decode(base64str string) string {
	if result, err := base64.StdEncoding.DecodeString(base64str); err == nil {
		return string(result)
	}
	return ""
}

// url编码
func UrlEncode(str string) string {
	return base64.URLEncoding.EncodeToString([]byte(str))
}

// url解码
func UrlDecode(str string) string {
	if result, err := base64.URLEncoding.DecodeString(str); err == nil {
		return string(result)
	}
	return ""
}
