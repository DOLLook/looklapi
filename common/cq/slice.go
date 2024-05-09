package cq

import (
	"slices"
)

// 泛型切片
type GenericSlice[T any] []T

func FromSlice[T any](slice []T) GenericSlice[T] {
	return slice
}

// 泛型切片使用selector生成泛型切片
func FromSliceSelect[T any, Out any](slice []T, selector func(e T) (Out, bool)) GenericSlice[Out] {
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

// 返回切片第一个满足条件的元素
func (source GenericSlice[T]) First(f func(e T) bool) (T, bool) {
	var first T
	if len(source) <= 0 {
		return first, false
	}

	if f == nil {
		return source[0], true
	}

	for _, item := range source {
		if f(item) {
			return item, true
		}
	}

	return first, false
}

// 返回切片第一个满足条件的元素
func (source GenericSlice[T]) FirstOrDefault(f func(e T) bool, defaultValue T) T {
	if len(source) <= 0 {
		return defaultValue
	}

	if f == nil {
		return source[0]
	}

	for _, item := range source {
		if f(item) {
			return item
		}
	}

	return defaultValue
}

// 返回切片最后一个满足条件的元素
func (source GenericSlice[T]) Last(f func(e T) bool) (T, bool) {
	var last T
	length := len(source)
	if length <= 0 {
		return last, false
	}

	if f == nil {
		return source[length-1], true
	}

	for i := length - 1; i >= 0; i-- {
		item := source[i]
		if f(item) {
			return item, true
		}
	}

	return last, false
}

// 返回切片最后一个满足条件的元素
func (source GenericSlice[T]) LastOrDefault(f func(e T) bool, defaultValue T) T {
	length := len(source)
	if length <= 0 {
		return defaultValue
	}

	if f == nil {
		return source[length-1]
	}

	for i := length - 1; i >= 0; i-- {
		item := source[i]
		if f(item) {
			return item
		}
	}

	return defaultValue
}

// 检测切片元素是否全部满足测试条件f
func (source GenericSlice[T]) All(f func(e T) bool) bool {
	if len(source) <= 0 || f == nil {
		return false
	}

	for _, item := range source {
		if !f(item) {
			return false
		}
	}

	return true
}

// 检测切片是否存在元素满足测试条件f
func (source GenericSlice[T]) Any(f func(e T) bool) bool {
	if len(source) <= 0 || f == nil {
		return false
	}

	for _, item := range source {
		if f(item) {
			return true
		}
	}

	return false
}

// 泛型切片过滤器
func (source GenericSlice[T]) Filter(f func(e T) bool) GenericSlice[T] {
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

// 泛型切片排序
// 返回排序后新切片 不改变原始切片顺序
func (source GenericSlice[T]) Sort(cmp func(a, b T) int) GenericSlice[T] {
	var sorted []T
	if len(source) <= 0 {
		return sorted
	}

	sorted = make([]T, len(source))
	copy(sorted, source)
	if cmp == nil {
		return sorted
	}

	slices.SortFunc(sorted, cmp)

	return sorted
}

// 泛型切片排序 保持相同元素顺序
// 返回排序后新切片 不改变原始切片顺序
func (source GenericSlice[T]) SortStable(cmp func(a, b T) int) GenericSlice[T] {
	var sorted []T
	if len(source) <= 0 {
		return sorted
	}

	sorted = make([]T, len(source))
	copy(sorted, source)
	if cmp == nil {
		return sorted
	}

	slices.SortStableFunc(sorted, cmp)

	return sorted
}

// 泛型切片迭代
func (source GenericSlice[T]) Foreach(f func(e T)) {
	if len(source) <= 0 || f == nil {
		return
	}

	for _, item := range source {
		f(item)
	}
}

func (source GenericSlice[T]) ToSlice() []T {
	return source
}

func (source GenericSlice[T]) Length() int {
	return len(source)
}

// 切片计数
func (source GenericSlice[T]) Count(f func(e T) bool) int {
	if f == nil {
		return len(source)
	}

	count := 0
	for _, item := range source {
		if f(item) {
			count++
		}
	}

	return count
}

// 切片去重
func (source GenericSlice[T]) Distinct(distinctBy func(e T) any) GenericSlice[T] {
	var distinct []T

	if len(source) < 1 {
		return distinct
	}

	if distinctBy == nil {
		m := make(map[any]bool, len(source))
		for _, item := range source {
			if _, has := m[item]; !has {
				distinct = append(distinct, item)
				m[item] = true
			}
		}

	} else {
		m := make(map[any]bool, len(source))
		for _, item := range source {
			k := distinctBy(item)
			if _, has := m[k]; !has {
				distinct = append(distinct, item)
				m[k] = true
			}
		}
	}

	return distinct
}

// 返回切片唯一满足条件的元素
func (source GenericSlice[T]) Single(f func(e T) bool) (T, bool) {
	var single T
	if len(source) <= 0 {
		return single, false
	}

	if f == nil {
		if len(source) == 1 {
			return source[0], true
		} else {
			return single, false
		}
	}

	count := 0
	for _, item := range source {
		if f(item) {
			if count > 0 {
				var zero T
				return zero, false
			}
			single = item
			count++
		}
	}

	return single, count > 0
}
