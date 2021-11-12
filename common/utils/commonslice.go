package utils

import (
	"go-webapi-fw/errs"
	"reflect"
)

// 通用切片
type commonSlice []interface{}

func NewCommonSlice(slice interface{}) commonSlice {
	if slice == nil {
		return nil
	}

	sval := reflect.ValueOf(slice)
	if sval.Kind() != reflect.Slice {
		panic(errs.NewBllError("unvalid slice"))
	}

	slen := sval.Len()
	commonsl := make(commonSlice, slen)
	for i := 0; i < slen; i++ {
		commonsl[i] = sval.Index(i).Interface()
	}

	return commonsl
}

// 过滤器
func (source commonSlice) Filter(f func(item interface{}) bool) commonSlice {
	if source == nil || len(source) == 0 {
		return source
	}

	tempSlice := commonSlice{}
	for _, item := range source {
		if f(item) {
			tempSlice = append(tempSlice, item)
		}
	}

	return tempSlice
}
