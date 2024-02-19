package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
)

// MD5签名算法
//
// @data 签名map
//
// @suffixKey 指定后缀名称, 与suffixVal必须同时为空或不为空
//
// @suffixVal 指定后缀值, 与suffixKey必须同时为空或不为空
func MD5Sign(data map[string]string, suffixKey string, suffixVal string) (string, error) {
	if len(data) < 1 {
		return "", nil
	}

	if (IsEmpty(suffixKey) && !IsEmpty(suffixVal)) || (!IsEmpty(suffixKey) && IsEmpty(suffixVal)) {
		return "", errors.New("suffixKey and suffixVal must be all empty or not empty")
	}

	var keys []string
	for k, _ := range data {
		if IsEmpty(k) {
			return "", errors.New("key in data to sign can not be empty")
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)
	url := ""
	for i, k := range keys {
		val := data[k]
		if IsEmpty(val) {
			return "", errors.New("value in data to sign can not be empty")
		}
		if i == 0 {
			url = url + k + "=" + val
		} else {
			url = url + "&" + k + "=" + val
		}
	}

	if !IsEmpty(suffixKey) {
		url = url + "&" + suffixKey + "=" + suffixVal
	}

	_md5 := MD5(url)
	return strings.ToUpper(_md5), nil
}

// SHA1Hash 对字符串进行sha1签名
func SHA1Hash(targetString string) string {
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(targetString))
	result := sha1Hash.Sum(nil)
	return hex.EncodeToString(result)
}

// 获取未使用urlEncode的form data
func GetFormDataWithoutEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}
