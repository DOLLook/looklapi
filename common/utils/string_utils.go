package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// 获取字符串MD5 bytes
func MD5Bytes(str string) []byte {
	_md5 := md5.New()
	_md5.Write([]byte(str))
	return _md5.Sum([]byte(nil))
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

// base64 url编码
func Base64UrlEncode(str string) string {
	return base64.URLEncoding.EncodeToString([]byte(str))
}

// base64 url解码
func Base64UrlDecode(str string) string {
	if result, err := base64.URLEncoding.DecodeString(str); err == nil {
		return string(result)
	}
	return ""
}

// url编码
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// url解码
func UrlDecode(str string) string {
	if result, err := url.QueryUnescape(str); err == nil {
		return result
	}
	return ""
}

// 生成随机字符串
func GetRandomString(codeLen int) string {
	if codeLen <= 0 {
		return ""
	}
	rd := rand.New(rand.NewSource(time.Now().UnixMilli()))
	//rand.Seed(time.Now().UnixMilli())
	bytes := make([]byte, codeLen)
	//rand.Read(bytes)
	rd.Read(bytes)
	return hex.EncodeToString(bytes)
}

func InterfaceToString(v interface{}) string {
	var result = ""
	if v != nil {
		switch v.(type) {
		case float64:
			ft := v.(float64)
			result = strconv.FormatFloat(ft, 'f', -1, 64)
		case float32:
			ft := v.(float32)
			result = strconv.FormatFloat(float64(ft), 'f', -1, 64)
		case int:
			it := v.(int)
			result = strconv.Itoa(it)
		case uint:
			it := v.(uint)
			result = strconv.Itoa(int(it))
		case int8:
			it := v.(int8)
			result = strconv.Itoa(int(it))
		case uint8:
			it := v.(uint8)
			result = strconv.Itoa(int(it))
		case int16:
			it := v.(int16)
			result = strconv.Itoa(int(it))
		case uint16:
			it := v.(uint16)
			result = strconv.Itoa(int(it))
		case int32:
			it := v.(int32)
			result = strconv.Itoa(int(it))
		case uint32:
			it := v.(uint32)
			result = strconv.Itoa(int(it))
		case int64:
			it := v.(int64)
			result = strconv.FormatInt(it, 10)
		case uint64:
			it := v.(uint64)
			result = strconv.FormatUint(it, 10)
		case string:
			result = v.(string)
		case []byte:
			result = string(v.([]byte))
		default:
			newValue, _ := json.Marshal(v)
			result = string(newValue)
		}
	}
	return result
}

// 获取随机字符串 size:字符串长度 kind:2 纯数字, 4 小写字母, 8 大写字母, 可|运算组合
func RandomString(size int, kind int) string {
	if size < 1 {
		return ""
	}

	kinds := [][]int{
		[]int{10, 48}, // 数字
		[]int{26, 97}, // 小写字母
		[]int{26, 65}, // 大写字母
	}
	rsBytes := make([]byte, size)

	iKind := -1
	kindIndexPool := make([]int, 0)
	if kind&2 == 2 {
		iKind = 0
		kindIndexPool = append(kindIndexPool, 0)
	}
	if kind&4 == 4 {
		iKind = 1
		kindIndexPool = append(kindIndexPool, 1)
	}
	if kind&8 == 8 {
		iKind = 2
		kindIndexPool = append(kindIndexPool, 2)
	}

	if iKind < 0 {
		kindIndexPool = append(kindIndexPool, 0, 1, 2)
	}
	kindIndexPoolLen := len(kindIndexPool)

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < size; i++ {
		if kindIndexPoolLen > 1 {
			iKind = kindIndexPool[rd.Intn(kindIndexPoolLen)]
		}
		scope, base := kinds[iKind][0], kinds[iKind][1]
		rsBytes[i] = uint8(base + rd.Intn(scope))
	}
	return string(rsBytes)
}
