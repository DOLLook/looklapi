package redisutils

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"looklapi/common/utils"
	"reflect"
	"strings"
	"time"
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

func HashGetValues(key string, hashFields interface{}, slicePtr interface{}) error {
	if utils.IsEmpty(key) || hashFields == nil {
		return errors.New("invalid key or hashFields")
	}

	if slicePtr == nil {
		return errors.New("slicePtr must not be nil")
	}

	sliceValRef := reflect.ValueOf(slicePtr)
	if sliceValRef.Kind() != reflect.Ptr || sliceValRef.Elem().Kind() != reflect.Slice {
		return errors.New("slicePtr must be a slice pointer")
	}

	fileds := reflect.ValueOf(hashFields)
	if fileds.Kind() != reflect.Slice {
		return errors.New("hashFields must be a hash filed slice")
	}

	keys := make([]interface{}, 0)
	for i := 0; i < fileds.Len(); i++ {
		keys = append(keys, fileds.Index(i).Interface())
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	cmd := fmt.Sprintf("local rst={}; for i,v in pairs(KEYS) do rst[i]=redis.call('hget','%s',v) end; return rst;", key)
	reply, err := redis.NewScript(len(keys), cmd).Do(conn, keys...)
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

func HashKeys(key string, valPtr interface{}) error {
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

	reply, err := conn.Do("HKEYS", key)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

func HashValues(key string, slicePtr interface{}) error {
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

	reply, err := conn.Do("HVALS", key)
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

// 设置key过期时间 expSecs(秒)
func SetKeyExpSecs(key string, expSecs int) error {
	if utils.IsEmpty(key) || expSecs < 1 {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expSecs)

	return err
}

// 设置key过期时间 expMillSecs(毫秒)
func SetKeyExpMillSecs(key string, expMillSecs int) error {
	if utils.IsEmpty(key) || expMillSecs < 1 {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("PEXPIRE", key, expMillSecs)

	return err
}

// 设置key过期时间 按给定expTime以秒为单位的时间戳
func SetKeyExpUnixSecs(key string, expTime time.Time) error {
	if utils.IsEmpty(key) || expTime.IsZero() {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("EXPIREAT", key, expTime.Unix())

	return err
}

// 设置key过期时间 按给定expTime以毫秒为单位的时间戳
func SetKeyExpUnixMillSecs(key string, expTime time.Time) error {
	if utils.IsEmpty(key) || expTime.IsZero() {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("PEXPIREAT", key, expTime.UnixMilli())

	return err
}

// 持久化key 将key的过期时间移除
func RemoveKeyExp(key string) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("PERSIST", key)

	return err
}

// 获取key剩余存活秒数 当 key不存在时，返回 -2 。 当key存在但没有设置剩余生存时间时，返回 -1
func GetKeyTimeToLiveSecs(key string, secPtr *int64) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("TTL", key)
	if err != nil {
		return err
	}

	return parse(reply, secPtr)
}

// 获取key剩余存活毫秒数
func GetKeyTimeToLiveMillSecs(key string, millSecPtr *int64) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("PTTL", key)
	if err != nil {
		return err
	}

	return parse(reply, millSecPtr)
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
