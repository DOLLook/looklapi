package cq

import "golang.org/x/exp/maps"

// comparable泛型切片
type ComparableSlice[T comparable] []T

func FromComparableSlice[T comparable](slice []T) ComparableSlice[T] {
	return slice
}

// 泛型切片使用selector生成comparable泛型切片
func FromSliceSelectComparable[T any, Out comparable](slice []T, selector func(e T) (Out, bool)) ComparableSlice[Out] {
	var out []Out
	if len(slice) < 1 || selector == nil {
		return out
	}
	for _, item := range slice {
		if o, ok := selector(item); ok {
			out = append(out, o)
		}
	}
	return out
}

func (source ComparableSlice[T]) ToGenericSlice() GenericSlice[T] {
	return GenericSlice[T](source)
}

func (source ComparableSlice[T]) ToSlice() []T {
	return source
}

func (source ComparableSlice[T]) Length() int {
	return len(source)
}

// comparable泛型切片去重
func (source ComparableSlice[T]) Distinct() ComparableSlice[T] {
	var distinct []T

	if len(source) < 1 {
		return distinct
	}

	m := make(map[T]bool, len(source))
	for _, item := range source {
		if _, has := m[item]; !has {
			distinct = append(distinct, item)
			m[item] = true
		}
	}

	return distinct
}

// comparable泛型切片排除集合
func (source ComparableSlice[T]) Except(except []T) ComparableSlice[T] {
	var excepted []T
	if len(source) < 1 {
		return excepted
	}

	if len(except) < 1 {
		return source
	}

	m := make(map[T]bool, len(except))
	for _, item := range except {
		m[item] = true
	}

	for _, item := range source {
		if _, has := m[item]; !has {
			excepted = append(excepted, item)
		}
	}

	return excepted
}

// comparable泛型切片求交集
func (source ComparableSlice[T]) Intersect(slice []T) ComparableSlice[T] {
	var intersect []T
	rightLen := len(slice)
	if rightLen < 1 {
		return intersect
	}
	leftLen := len(source)
	if leftLen < 1 {
		return intersect
	}

	if leftLen < rightLen {
		rightChecked := 0
		tempMap := make(map[T]bool)
		for _, lItem := range source {
			if _, has := tempMap[lItem]; has {
				intersect = append(intersect, lItem)
				continue
			}

			for i := rightChecked; i < rightLen; i++ {
				rightChecked = i + 1
				rItem := slice[i]
				tempMap[rItem] = true
				if lItem == rItem {
					intersect = append(intersect, lItem)
					break
				}
			}
		}
	} else {
		leftChecked := 0
		tempMap := make(map[T]bool)
		for _, rItem := range slice {
			if _, has := tempMap[rItem]; has {
				intersect = append(intersect, rItem)
				continue
			}

			for i := leftChecked; i < leftLen; i++ {
				leftChecked = i + 1
				lItem := source[i]
				tempMap[lItem] = true
				if rItem == lItem {
					intersect = append(intersect, rItem)
					break
				}
			}
		}
	}

	if len(intersect) <= 0 {
		return intersect
	}

	distinct := maps.Keys(SliceToMap(intersect, func(e T) T {
		return e
	}, func(e T) bool {
		return true
	}))

	return distinct
}
