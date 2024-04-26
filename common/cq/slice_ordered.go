package cq

import (
	"cmp"
	"slices"
)

// ordered泛型切片
type OrderedSlice[T cmp.Ordered] []T

func FromOrderedSlice[T cmp.Ordered](slice []T) OrderedSlice[T] {
	return slice
}

// 泛型切片使用selector生成ordered泛型切片
func FromSliceSelectOrdered[T any, Out cmp.Ordered](slice []T, selector func(e T) (Out, bool)) OrderedSlice[Out] {
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
