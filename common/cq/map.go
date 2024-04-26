package cq

// 泛型map
type GenericMap[K comparable, V any] map[K]V

func FromMap[K comparable, V any](m map[K]V) GenericMap[K, V] {
	return m
}

// 泛型map过滤器
func (source GenericMap[K, V]) FilterMap(filterK func(key K) bool, filterV func(val V) bool) GenericMap[K, V] {
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
func (source GenericMap[K, V]) FilterKeys(filterK func(key K) bool, filterV func(val V) bool) GenericSlice[K] {
	var keys []K

	if len(source) <= 0 {
		return keys
	}

	if filterK == nil && filterV == nil {
		keys = make([]K, len(source))
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
func (source GenericMap[K, V]) FilterValues(filterK func(key K) bool, filterV func(val V) bool) GenericSlice[V] {
	var values []V

	if len(source) <= 0 {
		return values
	}

	if filterK == nil && filterV == nil {
		values = make([]V, len(source))
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

// 泛型map计数
func (source GenericMap[K, V]) Count(filterK func(key K) bool, filterV func(val V) bool) int {
	sLen := len(source)
	if sLen <= 0 {
		return sLen
	}

	if filterK == nil && filterV == nil {
		return sLen
	}

	count := 0
	if filterK == nil {
		for _, v := range source {
			if filterV(v) {
				count++
			}
		}
	} else if filterV == nil {
		for k, _ := range source {
			if filterK(k) {
				count++
			}
		}
	} else {
		for k, v := range source {
			if filterK(k) && filterV(v) {
				count++
			}
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
