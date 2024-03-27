package collection_utils

import (
	"cmp"
	"golang.org/x/exp/maps"
	"slices"
)

// ordered泛型切片
type OrderedSlice[T cmp.Ordered] []T

// comparable泛型切片
type ComparableSlice[T comparable] []T

// 泛型切片
type GenericSlice[T any] []T

func FromOrderedSlice[T cmp.Ordered](slice []T) OrderedSlice[T] {
	return slice
}

// 泛型切片使用selector生成ordered泛型切片
func FromSliceOrderedSelect[T any, Out cmp.Ordered](slice []T, selector func(e T) (Out, bool)) OrderedSlice[Out] {
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

func FromComparableSlice[T comparable](slice []T) ComparableSlice[T] {
	return slice
}

// 泛型切片使用selector生成comparable泛型切片
func FromSliceComparableSelect[T any, Out comparable](slice []T, selector func(e T) (Out, bool)) ComparableSlice[Out] {
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
func (source GenericSlice[T]) First(f func(item T) bool) (T, bool) {
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

// 返回切片最后一个满足条件的元素
func (source GenericSlice[T]) Last(f func(item T) bool) (T, bool) {
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

// 检测切片元素是否全部满足测试条件f
func (source GenericSlice[T]) All(f func(item T) bool) bool {
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
func (source GenericSlice[T]) Any(f func(item T) bool) bool {
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
func (source GenericSlice[T]) Filter(f func(item T) bool) GenericSlice[T] {
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
func (source GenericSlice[T]) Foreach(f func(item T)) {
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
func (source GenericSlice[T]) Count(f func(item T) bool) int {
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
func (source GenericSlice[T]) Distinct(distinctBy func(item T) any) GenericSlice[T] {
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

func (source ComparableSlice[T]) Length() int {
	return len(source)
}

func (source ComparableSlice[T]) ToGenericSlice() GenericSlice[T] {
	return GenericSlice[T](source)
}

func (source ComparableSlice[T]) ToSlice() []T {
	return source
}

func (source OrderedSlice[T]) ToGenericSlice() GenericSlice[T] {
	return GenericSlice[T](source)
}

func (source OrderedSlice[T]) ToSlice() []T {
	return source
}

func (source OrderedSlice[T]) Length() int {
	return len(source)
}

// ordered泛型切片去重
func (source OrderedSlice[T]) Distinct() OrderedSlice[T] {
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

// ordered泛型slice求合
func (source OrderedSlice[T]) Sum() T {
	var sum T
	for _, item := range source {
		sum += item
	}

	return sum
}

// ordered泛型切片最小
func (source OrderedSlice[T]) Min() T {
	return slices.Min(source)
}

// ordered泛型切片最大
func (source OrderedSlice[T]) Max() T {
	return slices.Max(source)
}
