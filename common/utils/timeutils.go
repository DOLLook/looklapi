package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var _MinDateTime *time.Time

// 获取最小时间
func MinDateTime() time.Time {
	if _MinDateTime == nil {
		ntime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		_MinDateTime = &ntime
	}

	return *_MinDateTime
}

// 时间格式化 2006-01-02 15:04:05
func TimeFormat(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

// 时间格式化 "'2006-01-02 15:04:05'"
func SqlTimeFormat(time time.Time) string {
	return fmt.Sprintf("'%s'", time.Format("2006-01-02 15:04:05"))
}

// 字符串转本地时间 timeFormat为空时默认格式2006-01-02 15:04:05
func StringToTime(timeFormat string, timeStr string) (time.Time, error) {
	if IsEmpty(timeStr) {
		return time.Time{}, errors.New("invalid time string")
	}

	if IsEmpty(timeFormat) {
		timeFormat = "2006-01-02 15:04:05"
	}

	t, err := time.ParseInLocation(timeFormat, timeStr, time.Local)
	return t, err
}

// 时间简化格式 去除时区信息
// datetime 如2020-12-01T12:06:13Z  2020-12-01T12:06:13+08:00
// 简化为 2020-12-01 12:06:13
func DateTimeSimplifyTimezone(datetime string) string {
	datetime = strings.TrimSuffix(datetime, "Z")
	datetime = strings.ReplaceAll(datetime, "T", " ")
	datetime = strings.Split(datetime, "+")[0]
	datetime = strings.Split(datetime, "Z")[0]
	datetime = strings.TrimSpace(datetime)

	return datetime
}
