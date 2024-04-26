package cq

import (
	"github.com/shopspring/decimal"
	"reflect"
	"unsafe"
)

// decimal切片
type DecimalSlice GenericSlice[decimal.Decimal]

func (source DecimalSlice) ToGenericSlice() GenericSlice[decimal.Decimal] {
	return FromSlice(source)
}

func (source DecimalSlice) ToComparableSlice() ComparableSlice[decimal.Decimal] {
	return FromComparableSlice(source)
}

func (source DecimalSlice) ToSlice() []decimal.Decimal {
	return source
}

func (source DecimalSlice) Length() int {
	return len(source)
}

// decimal泛型切片转DecimalSlice
func (source GenericSlice[T]) ToDecimalSlice() DecimalSlice {
	if len(source) < 1 {
		return nil
	}
	v := reflect.ValueOf(source)
	slice := unsafe.Slice((*decimal.Decimal)(v.UnsafePointer()), len(source))
	//slice := unsafe.Slice((*decimal.Decimal)(unsafe.Pointer(&source[0])), len(source))
	return slice

	//return v.Convert(reflect.TypeOf((*DecimalSlice)(nil)).Elem()).Interface().(DecimalSlice)
}

// decimal泛型切片转DecimalSlice
func (source ComparableSlice[T]) ToDecimalSlice() DecimalSlice {
	if len(source) < 1 {
		return nil
	}
	v := reflect.ValueOf(source)
	slice := unsafe.Slice((*decimal.Decimal)(v.UnsafePointer()), len(source))
	return slice
}

// 求和
func (source DecimalSlice) Sum() decimal.Decimal {
	sum := decimal.Zero
	for _, item := range source {
		sum = sum.Add(item)
	}
	return sum
}

// 取最小
func (source DecimalSlice) Min() decimal.Decimal {
	l := len(source)
	if l < 1 {
		return decimal.Zero
	}

	ds := source[0]
	for i := 1; i < l; i++ {
		item := source[i]
		if item.LessThan(ds) {
			ds = item
		}
	}

	return ds
}

// 取最大
func (source DecimalSlice) Max() decimal.Decimal {
	l := len(source)
	if l < 1 {
		return decimal.Zero
	}

	ds := source[0]
	for i := 1; i < l; i++ {
		item := source[i]
		if item.GreaterThan(ds) {
			ds = item
		}
	}

	return ds
}
