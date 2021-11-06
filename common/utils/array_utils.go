package utils

import "reflect"

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

// 判断数组是否包含元素
// arrValueOf 数组运行时对象值
// ele 要检查的值
func ArrayContains(array interface{}, ele interface{}) bool {
	if array == nil {
		return false
	}

	arrValueOf := reflect.ValueOf(array)
	switch arrValueOf.Kind() {
	case reflect.Array, reflect.Slice:
		break
	default:
		panic("invalid array")
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
