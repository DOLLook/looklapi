package redisutils

import (
	"github.com/garyburd/redigo/redis"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/errs"
	"reflect"
)

func Set(key string, val interface{}) {
	if utils.IsEmpty(key) || val == nil {
		return
	}

	val = objConvertToJson(val)

	conn := getConn(key)
	_, err := conn.Do("SET", key, val)

	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}
}

func Get(key string, valPtr interface{}) {
	if utils.IsEmpty(key) {
		return
	}

	if valPtr == nil {
		panic(errs.NewBllError("valPtr must not be nil"))
	}

	conn := getConn(key)
	reply, err := conn.Do("GET", key)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	parse(reply, valPtr)
}

func HashSet(key string, hashField interface{}, val interface{}) {
	if utils.IsEmpty(key) || hashField == nil || val == nil {
		return
	}

	val = objConvertToJson(val)

	conn := getConn(key)
	_, err := conn.Do("HSET", key, hashField, val)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}
}

func HashGet(key string, hashField interface{}, valPtr interface{}) {
	if utils.IsEmpty(key) || hashField == nil {
		return
	}

	if valPtr == nil {
		panic(errs.NewBllError("valPtr must not be nil"))
	}

	conn := getConn(key)
	reply, err := conn.Do("HGET", key, hashField)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	parse(reply, valPtr)
}

// 判断key是否存在
func Exist(key string) bool {
	if utils.IsEmpty(key) {
		return false
	}

	conn := getConn(key)
	reply, err := conn.Do("EXISTS", key)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	var result bool
	parse(reply, &result)
	return result
}

// hash是否存在
func HExist(key string, hashField interface{}) bool {
	if utils.IsEmpty(key) || hashField == nil {
		return false
	}

	conn := getConn(key)
	reply, err := conn.Do("HEXISTS", key, hashField)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	var result bool
	parse(reply, &result)
	return result
}

// 删除key
func Del(key string) {
	if utils.IsEmpty(key) {
		return
	}

	conn := getConn(key)
	_, err := conn.Do("DEL", key)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}
}

// 删除key
func HDel(key string, hashField interface{}) {
	if utils.IsEmpty(key) || hashField == nil {
		return
	}

	conn := getConn(key)
	_, err := conn.Do("HDEL", key, hashField)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}
}

func parse(reply interface{}, valPtr interface{}) {
	objVale := reflect.ValueOf(valPtr).Elem()
	var valueInterface interface{}
	var parseErr error
	switch objVale.Kind() {
	case reflect.Bool:
		valueInterface, parseErr = redis.Bool(reply, nil)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		valueInterface, parseErr = redis.Int(reply, nil)
		break
	case reflect.Int64:
		valueInterface, parseErr = redis.Int64(reply, nil)
		break
	case reflect.Uint64:
		valueInterface, parseErr = redis.Uint64(reply, nil)
		break
	case reflect.Float32, reflect.Float64:
		valueInterface, parseErr = redis.Float64(reply, nil)
		break
	case reflect.String:
		valueInterface, parseErr = redis.String(reply, nil)
		break
	default:
		bytes := reply.([]byte)
		valStr := string(bytes)
		parseErr = utils.JsonToStruct(valStr, valPtr)
		break
	}

	if parseErr != nil {
		switch reply.(type) {
		case nil, redis.Error:
			break
		default:
			panic(errs.NewBllError(parseErr.Error()))
		}
	}

	if valueInterface != nil {
		resultValue := reflect.ValueOf(valueInterface)
		objVale.Set(resultValue)
	}
}

// 对象转json
func objConvertToJson(val interface{}) interface{} {
	reflectVal := reflect.ValueOf(val)
	kind := reflectVal.Kind()
	switch kind {
	case reflect.Interface, reflect.Ptr, reflect.Uintptr:
		kind = reflectVal.Elem().Kind()
		break
	default:
		break
	}

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		val = utils.StructToJson(val)
		break
	default:
		break
	}

	return val
}
