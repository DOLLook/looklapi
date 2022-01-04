package utils

import "time"

var _MinDateTime *time.Time

// 获取最小时间
func MinDateTime() time.Time {
	if _MinDateTime == nil {
		ntime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		_MinDateTime = &ntime
	}

	return *_MinDateTime
}
