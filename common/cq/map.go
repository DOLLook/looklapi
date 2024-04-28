package cq

import "golang.org/x/exp/maps"

// 泛型map
type GenericMap[K comparable, V any] map[K]V

func FromMap[K comparable, V any](m map[K]V) GenericMap[K, V] {
	return m
}

// 泛型map过滤器
func (source GenericMap[K, V]) FilterMap(predicate func(k K, v V) bool) GenericMap[K, V] {
	m := make(map[K]V)

	sLen := len(source)
	if sLen <= 0 {
		return m
	}

	if predicate == nil {
		m := make(map[K]V, sLen)
		for k, v := range source {
			m[k] = v
		}

		return m
	}

	for k, v := range source {
		if predicate(k, v) {
			m[k] = v
		}
	}

	return m
}

// 泛型map key过滤器
func (source GenericMap[K, V]) FilterKeys(predicate func(k K, v V) bool) GenericSlice[K] {
	var keys []K

	if len(source) <= 0 {
		return keys
	}

	if predicate == nil {
		return maps.Keys(source)
	}

	for k, v := range source {
		if predicate(k, v) {
			keys = append(keys, k)
		}
	}

	return keys
}

// 泛型map value过滤器
func (source GenericMap[K, V]) FilterValues(predicate func(k K, v V) bool) GenericSlice[V] {
	var values []V

	if len(source) <= 0 {
		return values
	}

	if predicate == nil {
		return maps.Values(source)
	}

	for k, v := range source {
		if predicate(k, v) {
			values = append(values, v)
		}
	}

	return values
}

// 泛型map计数
func (source GenericMap[K, V]) Count(predicate func(k K, v V) bool) int {
	sLen := len(source)
	if sLen <= 0 {
		return sLen
	}

	if predicate == nil {
		return sLen
	}

	count := 0
	for k, v := range source {
		if predicate(k, v) {
			count++
		}
	}

	return count
}

func (source GenericMap[K, V]) Length() int {
	return len(source)
}

func (source GenericMap[K, V]) ToMap() map[K]V {
	return source
}

// 检测map是否存在k/v满足测试条件predicate
func (source GenericMap[K, V]) Any(predicate func(k K, v V) bool) bool {
	if source == nil || len(source) < 1 {
		return false
	}

	if predicate == nil {
		return true
	}

	for k, v := range source {
		if predicate(k, v) {
			return true
		}
	}

	return false
}
