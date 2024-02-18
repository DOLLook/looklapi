package utils

import (
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
		return nil
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

// 泛型切片
type GenericSlice[T any] []T

// 泛型map
type GenericMap[K comparable, V any] map[K]V

func FromSlice[T any](slice []T) GenericSlice[T] {
	return slice
}

func FromMap[K comparable, V any](m map[K]V) GenericMap[K, V] {
	return m
}

// 泛型切片过滤器
func (source GenericSlice[T]) FilterT(f func(item T) bool) []T {
	var temp []T
	if len(source) <= 0 {
		return temp
	}

	if f == nil {
		temp = make([]T, len(source))
		copy(temp, source)
		return temp
	}

	for _, item := range source {
		if f(item) {
			temp = append(temp, item)
		}
	}

	return temp
}

// 泛型map过滤器
func (source GenericMap[K, V]) FilterMap(filterK func(key K) bool, filterV func(val V) bool) map[K]V {
	m := make(map[K]V)

	sLen := len(source)
	if sLen <= 0 {
		return m
	}

	if filterK == nil && filterV == nil {
		m := make(map[K]V, sLen)
		for k, v := range source {
			m[k] = v
		}

		return m
	}

	if filterK == nil {
		for k, v := range source {
			if filterV(v) {
				m[k] = v
			}
		}
	} else if filterV == nil {
		for k, v := range source {
			if filterK(k) {
				m[k] = v
			}
		}
	} else {
		for k, v := range source {
			if filterK(k) && filterV(v) {
				m[k] = v
			}
		}
	}

	return m
}

// 泛型map key过滤器
func (source GenericMap[K, V]) FilterKeys(filterK func(key K) bool, filterV func(val V) bool) []K {
	var keys []K

	if len(source) <= 0 {
		return keys
	}

	if filterK == nil && filterV == nil {
		keys := make([]K, len(source))
		i := 0
		for k, _ := range source {
			keys[i] = k
			i++
		}
	} else if filterK == nil {
		for k, v := range source {
			if filterV(v) {
				keys = append(keys, k)
			}
		}
	} else if filterV == nil {
		for k, _ := range source {
			if filterK(k) {
				keys = append(keys, k)
			}
		}
	} else {
		for k, v := range source {
			if filterK(k) && filterV(v) {
				keys = append(keys, k)
			}
		}
	}

	return keys
}

// 泛型map value过滤器
func (source GenericMap[K, V]) FilterValues(filterK func(key K) bool, filterV func(val V) bool) []V {
	var values []V

	if len(source) <= 0 {
		return values
	}

	if filterK == nil && filterV == nil {
		values := make([]V, len(source))
		i := 0
		for _, v := range source {
			values[i] = v
			i++
		}
	} else if filterK == nil {
		for _, v := range source {
			if filterV(v) {
				values = append(values, v)
			}
		}
	} else if filterV == nil {
		for k, v := range source {
			if filterK(k) {
				values = append(values, v)
			}
		}
	} else {
		for k, v := range source {
			if filterK(k) && filterV(v) {
				values = append(values, v)
			}
		}
	}

	return values
}
