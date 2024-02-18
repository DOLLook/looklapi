package redisutils

import (
	"errors"
	"fmt"
	"looklapi/common/utils"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// list: 左侧为头, 右侧为尾

// 向列表头(左端)push数据 多个值为原子push
// 返回push后列表的长度lenAfterPush
func LPush(key string, lenAfterPush *int, values ...interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if lenAfterPush != nil && *lenAfterPush >= 0 {
		return errors.New("the lenAfterPush must init to less than 0")
	}

	if len(values) < 1 {
		return nil
	}

	objs := make([]interface{}, 0)
	objs = append(objs, key)

	for _, item := range values {
		if item == nil {
			return errors.New("value that push to list can not be nil")
		}
		temp := objConvertToJson(item)
		objs = append(objs, temp)
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LPUSH", objs...)
	if err != nil {
		return err
	}

	if lenAfterPush == nil {
		return nil
	}

	return parse(reply, lenAfterPush)
}

// 向列表尾(右端)push数据 多个值为原子push
// 返回push后列表的长度lenAfterPush
func RPush(key string, lenAfterPush *int, values ...interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if lenAfterPush != nil && *lenAfterPush >= 0 {
		return errors.New("the lenAfterPush must init to less than 0")
	}

	if len(values) < 1 {
		return nil
	}

	objs := make([]interface{}, 0)
	objs = append(objs, key)

	for _, item := range values {
		if item == nil {
			return errors.New("value that push to list can not be nil")
		}
		temp := objConvertToJson(item)
		objs = append(objs, temp)
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("RPUSH", objs...)
	if err != nil {
		return err
	}

	if lenAfterPush == nil {
		return nil
	}

	return parse(reply, lenAfterPush)
}

// 移除并返回表头(左端)数据
func LPop(key string, valPtr interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LPOP", key)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 移除并返回表尾(右端)数据
func RPop(key string, valPtr interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("RPOP", key)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 在一个原子操作内 移除sourceKey的表尾(右端)数据sourceValue，且将sourceValue push到destinationKey的表头(左端)，并返回sourceValue
func RPopLPush(sourceKey, destinationKey string, valPtr interface{}) error {
	if utils.IsEmpty(sourceKey) || utils.IsEmpty(destinationKey) {
		return errors.New("invalid key")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	sourceDbIndex, destinationDbIndex := 0, 0
	sourceSlice := strings.Split(sourceKey, "_")
	if len(sourceSlice) > 1 {
		if reg, err := regexp.Compile("\\d+"); err == nil {
			match := reg.FindString(sourceSlice[0])
			if !utils.IsEmpty(match) {
				if dbIndex, err := strconv.Atoi(match); err == nil && dbIndex < _MAXDBINDEX {
					sourceDbIndex = dbIndex
				}
			}
		}
	}
	destinationSlice := strings.Split(destinationKey, "_")
	if len(destinationSlice) > 1 {
		if reg, err := regexp.Compile("\\d+"); err == nil {
			match := reg.FindString(destinationSlice[0])
			if !utils.IsEmpty(match) {
				if dbIndex, err := strconv.Atoi(match); err == nil && dbIndex < _MAXDBINDEX {
					destinationDbIndex = dbIndex
				}
			}
		}
	}

	if sourceDbIndex != destinationDbIndex {
		return errors.New("sourceKey and destinationKey must in the same db")
	}

	conn := getConn0(uint8(sourceDbIndex))
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("RPOPLPUSH", sourceKey, destinationKey)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 根据参数count的值，移除列表中与value相等的元素
// count > 0 : 从表头开始向表尾搜索，移除与 value 相等的元素，数量为 count 。
// count < 0 : 从表尾开始向表头搜索，移除与 value 相等的元素，数量为 count 的绝对值。
// count = 0 : 移除表中所有与 value 相等的值。
// 返回实际移除数量removeCount
func LRemove(key string, count int, value string, removeCount *int) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if removeCount != nil && *removeCount >= 0 {
		return errors.New("the removeCount must init to less than 0")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LREM", key, count, value)
	if err != nil {
		return err
	}

	if removeCount == nil {
		return nil
	}

	return parse(reply, removeCount)
}

// 返回列表 key 的长度
func LLen(key string, listLen *int) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if listLen == nil {
		return errors.New("need a none nil int pointer to receive the len")
	}

	if *listLen >= 0 {
		return errors.New("the listLen must init to less than 0")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LLEN", key)
	if err != nil {
		return err
	}

	return parse(reply, listLen)
}

// 返回列表 key 中，下标为 index 的元素 如果 index 参数的值不在列表的区间范围内(out of range)，返回 redis.ErrNil
func LIndex(key string, index int, valPtr interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LINDEX", key, index)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 将列表 key 下标为 index 的元素的值设置为 value
// 当 index 参数超出范围，或对一个空列表( key 不存在)进行 LSET 时，返回一个错误
func LSet(key string, index int, value interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if value == nil {
		return errors.New("value in list can not set nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("LSET", key, index, value)

	return err
}

// 返回列表 key 中指定区间[start,end]闭区间内的元素
// 超出范围的下标值不会引起错误
func LRange(key string, start int, end int, slicePtr interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if slicePtr == nil {
		return errors.New("slicePtr must not be nil")
	}

	sliceValRef := reflect.ValueOf(slicePtr)
	if sliceValRef.Kind() != reflect.Ptr || sliceValRef.Elem().Kind() != reflect.Slice {
		return errors.New("slicePtr must be a slice pointer")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("LRANGE", key, start, end)
	if err != nil {
		return err
	}

	sliceElem := sliceValRef.Elem().Type().Elem()
	if sliceElem.Kind() == reflect.Ptr {
		sliceElem = sliceElem.Elem()
	}

	if sliceElem.Kind() == reflect.Struct {
		// struct need use string to collect redis result, then use json to Unmarshal
		stringSlice := make([]string, 0)
		if err := parse(reply, &stringSlice); err != nil {
			return err
		}

		if len(stringSlice) > 0 {
			jsonStr := fmt.Sprintf("[%s]", strings.Join(stringSlice, ","))
			if err := utils.JsonToStruct(jsonStr, slicePtr); err != nil {
				return err
			} else {
				return nil
			}
		}

		return nil
	} else {
		return parse(reply, slicePtr)
	}
}
