package utils

import (
	"reflect"
)

// 检查字符串str是否包含checkStr
func StrContians(str string, checkStr string) bool {
	if IsEmpty(checkStr) {
		return true
	}

	if IsEmpty(str) {
		return false
	}

	strRunes, checkStrRunes := []rune(str), []rune(checkStr)
	lenStr, lenCheckStr := len(strRunes), len(checkStrRunes)

	if lenStr < lenCheckStr {
		return false
	}

	for i := 0; i < lenStr; i++ {
		tempStr := string(strRunes[i : i+lenCheckStr])
		if tempStr == checkStr {
			return true
		}
	}

	return false
}

// 判断数组或切片是否包含元素
// arrayOrSlice 数组或切片运行时对象值
// ele 要检查的值
func ArrayOrSliceContains(arrayOrSlice interface{}, ele interface{}) bool {
	if arrayOrSlice == nil {
		return false
	}

	arrValueOf := reflect.ValueOf(arrayOrSlice)
	switch arrValueOf.Kind() {
	case reflect.Array, reflect.Slice:
		break
	default:
		panic("invalid array or slice")
	}

	arrLen := arrValueOf.Len()
	if arrLen == 0 {
		return false
	}

	var notNil interface{}
	for i := 0; i < arrLen; i++ {
		item := arrValueOf.Index(i).Interface()
		if item == nil {
			if ele == nil {
				return true
			}
		} else {
			notNil = item
			break
		}
	}

	if notNil == nil {
		return ele == nil
	}

	itemValue := reflect.ValueOf(notNil)
	itemKind := itemValue.Kind()
	switch itemKind {
	case reflect.Interface, reflect.Ptr:
		itemKind = itemValue.Elem().Kind()
		break
	default:
		break
	}

	if itemKind == reflect.Struct {
		for i := 0; i < arrLen; i++ {
			item := arrValueOf.Index(i).Interface()
			if item == nil && ele == nil {
				return true
			}
			if item == ele {
				return true
			}
		}
	} else {
		if ele == nil {
			return false
		}
		for i := 0; i < arrLen; i++ {
			item := arrValueOf.Index(i).Interface()
			if item == ele {
				return true
			}
		}
	}

	return false
}

/**
删除切片元素
slicePtr 切片指针
firstCount 删除前几个, 0全部删除
返回 是否有值被移除
*/
func SliceRemove(slicePtr interface{}, ele interface{}, firstCount int) bool {
	if slicePtr == nil {
		return false
	}

	sValueOf := reflect.ValueOf(slicePtr).Elem()
	switch sValueOf.Kind() {
	case reflect.Slice:
		break
	default:
		panic("invalid slicePtr")
	}

	arrLen := sValueOf.Len()
	if arrLen == 0 {
		return false
	}

	itemType := reflect.TypeOf(slicePtr).Elem()
	itemKind := itemType.Elem().Kind()
	switch itemKind {
	case reflect.Interface, reflect.Ptr:
		itemKind = itemType.Elem().Kind()
		break
	default:
		break
	}

	if itemKind != reflect.Struct && ele == nil {
		return false
	}

	var removeMap = make(map[int]int, arrLen)
	var removeIndex []int

	if firstCount > 0 {
		// 仅删除前firstCount匹配项
		hasBreak := false
		tempLen := len(removeIndex)
		j := 0
		for i := 0; i < arrLen; i++ {
			item := sValueOf.Index(i).Interface()
			if !hasBreak && item == ele {
				if tempLen < firstCount {
					removeIndex = append(removeIndex, i-len(removeIndex))
					tempLen = len(removeIndex)
					if tempLen == firstCount {
						hasBreak = true
					}
					continue
				}
			}

			removeMap[j] = i
			j++
		}
	} else {
		// 删除所有匹配项
		j := 0
		for i := 0; i < arrLen; i++ {
			item := sValueOf.Index(i).Interface()
			if item == ele {
				removeIndex = append(removeIndex, i-len(removeIndex))
				continue
			}

			removeMap[j] = i
			j++
		}
	}

	removeLen := len(removeIndex)
	if removeLen == 0 {
		return false
	}

	for key, index := range removeMap {
		sValueOf.Index(key).Set(sValueOf.Index(index))
	}

	newLen := arrLen - removeLen
	sValueOf.SetLen(newLen)
	sValueOf.SetCap(newLen)

	return true
}

/**
删除切片元素
slicePtr 切片指针
index 待删除的索引
返回 是否有值被移除
*/
func SliceRemoveByIndex(slicePtr interface{}, index ...int) bool {
	if slicePtr == nil || len(index) == 0 {
		return false
	}

	sValueOf := reflect.ValueOf(slicePtr).Elem()
	switch sValueOf.Kind() {
	case reflect.Slice:
		break
	default:
		panic("invalid slicePtr")
	}

	arrLen := sValueOf.Len()
	if arrLen == 0 {
		return false
	}

	itemType := reflect.TypeOf(slicePtr).Elem()
	itemKind := itemType.Elem().Kind()
	switch itemKind {
	case reflect.Interface, reflect.Ptr:
		itemKind = itemType.Elem().Kind()
		break
	default:
		break
	}

	var removeMapCheck = make(map[int]bool, len(index))
	for _, i := range index {
		removeMapCheck[i] = false
	}

	change := false
	var removeMap = make(map[int]int, arrLen)
	// 删除所有匹配项
	j := 0
	for i := 0; i < arrLen; i++ {
		if hasRemove, ok := removeMapCheck[i]; ok {
			change = true
			if !hasRemove {
				removeMapCheck[i] = true
			}
			continue
		}

		removeMap[j] = i
		j++
	}

	if !change {
		return false
	}

	for key, index := range removeMap {
		sValueOf.Index(key).Set(sValueOf.Index(index))
	}

	newLen := arrLen - len(removeMapCheck)
	sValueOf.SetLen(newLen)
	sValueOf.SetCap(newLen)

	return true
}

/**
集合是否为空
*/
func CollectionIsEmpty(collection interface{}) bool {
	if collection == nil {
		return true
	}

	sValueOf := reflect.ValueOf(collection)
	switch sValueOf.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return sValueOf.Len() < 1
	case reflect.Ptr:
		switch sValueOf.Elem().Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return sValueOf.Elem().Len() < 1
		default:
			panic("invalid collection type")
		}
	default:
		panic("invalid collection type")
	}
}
