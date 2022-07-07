package redisutils

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"looklapi/common/utils"
	"reflect"
)

func Set(key string, val interface{}) error {
	if utils.IsEmpty(key) || val == nil {
		return errors.New("invalid arguments")
	}

	val = objConvertToJson(val)

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("SET", key, val)

	return err
}

func SetEx(key string, val interface{}, secs int) error {
	if utils.IsEmpty(key) || val == nil || secs < 1 {
		return errors.New("invalid arguments")
	}

	val = objConvertToJson(val)

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("SET", key, val, "EX", secs)

	return err
}

func Get(key string, valPtr interface{}) error {
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

	reply, err := conn.Do("GET", key)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

func HashSet(key string, hashField interface{}, val interface{}) error {
	if utils.IsEmpty(key) || hashField == nil || val == nil {
		return errors.New("invalid arguments")
	}

	val = objConvertToJson(val)

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("HSET", key, hashField, val)
	return err
}

func HashGet(key string, hashField interface{}, valPtr interface{}) error {
	if utils.IsEmpty(key) || hashField == nil {
		return errors.New("invalid key or hashField")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("HGET", key, hashField)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 判断key是否存在
func Exist(key string) (bool, error) {
	if utils.IsEmpty(key) {
		return false, nil
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return false, conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("EXISTS", key)
	if err != nil {
		return false, err
	}

	var result bool
	if err := parse(reply, &result); err != nil {
		return false, err
	}
	return result, nil
}

// hash是否存在
func HExist(key string, hashField interface{}) (bool, error) {
	if utils.IsEmpty(key) || hashField == nil {
		return false, nil
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return false, conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("HEXISTS", key, hashField)
	if err != nil {
		return false, err
	}

	var result bool
	if err := parse(reply, &result); err != nil {
		return false, err
	}
	return result, nil
}

// 删除key
func Del(key string) error {
	if utils.IsEmpty(key) {
		return nil
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

// 删除key
func HDel(key string, hashField interface{}) error {
	if utils.IsEmpty(key) || hashField == nil {
		return nil
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("HDEL", key, hashField)
	return err
}

func parse(reply interface{}, valPtr interface{}) (err error) {
	defer func() {
		if rcvErr := recover(); rcvErr != nil {
			switch rcvErr := rcvErr.(type) {
			case error:
				err = rcvErr
			case string:
				err = errors.New(rcvErr)
			default:
				err = errors.New("unknow err")
			}
		}
	}()

	switch reply := reply.(type) {
	case redis.Error:
		err = reply
	case nil:
		err = redis.ErrNil
	}
	if err != nil {
		return
	}

	objVale := reflect.ValueOf(valPtr).Elem()
	var valueInterface interface{}
	switch objVale.Kind() {
	case reflect.Bool:
		valueInterface, err = redis.Bool(reply, nil)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		valueInterface, err = redis.Int(reply, nil)
		break
	case reflect.Int64:
		valueInterface, err = redis.Int64(reply, nil)
		break
	case reflect.Uint64:
		valueInterface, err = redis.Uint64(reply, nil)
		break
	case reflect.Float32, reflect.Float64:
		valueInterface, err = redis.Float64(reply, nil)
		break
	case reflect.String:
		valueInterface, err = redis.String(reply, nil)
		break
	default:
		if bytes, ok := reply.([]byte); ok {
			valStr := string(bytes)
			err = utils.JsonToStruct(valStr, valPtr)
		} else if sl, ok := reply.([]interface{}); ok {
			err = redis.ScanSlice(sl, valPtr)
		} else {
			err = errors.New("invalid reply type")
		}
	}

	if err != nil {
		return
	}

	if valueInterface != nil {
		resultValue := reflect.ValueOf(valueInterface)
		objVale.Set(resultValue)
	}
	return
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
