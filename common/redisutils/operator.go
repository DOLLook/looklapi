package redisutils

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"looklapi/common/utils"
	"reflect"
	"strconv"
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

// 模糊查询keys
// dbIndex 数据库索引0-15
// pattern 模式匹配规则 示例: 前缀匹配 prefix*, 后缀匹配 *suffix, 中间匹配 *mid*
// limit 最大获取数量 0表示查询所有
func Scan(dbIndex uint8, pattern string, limit int) ([]string, error) {
	if dbIndex > 15 {
		return nil, errors.New("dbIndex must in [0,15]")
	}
	if utils.IsEmpty(pattern) {
		return nil, errors.New("pattern must not be empty")
	}
	if !strings.HasPrefix(pattern, "*") && !strings.HasSuffix(pattern, "*") {
		return nil, errors.New("pattern must be start with * or end with *")
	}

	conn := getConn0(dbIndex)
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	defer conn.Close()

	keys := make([]string, 0)
	cursor := "0"
	for {
		reply, err := conn.Do("SCAN", cursor, "MATCH", pattern)
		if err != nil {
			return nil, err
		}

		if rep, ok := reply.([]interface{}); ok && len(rep) == 2 {
			if ks, ok := rep[1].([]interface{}); ok && len(ks) > 0 {
				for _, item := range ks {
					k := string(item.([]byte))
					if !utils.IsEmpty(k) {
						keys = append(keys, k)
						if limit > 0 && len(keys) >= limit {
							break
						}
					}
				}
				if limit > 0 && len(keys) >= limit {
					break
				}
			}

			cursor = string(rep[0].([]byte))
			if cursor == "0" {
				break
			}
		} else {
			break
		}
	}

	return keys, nil
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

func HashMultiSet(key string, kv interface{}) error {
	if utils.IsEmpty(key) || kv == nil {
		return errors.New("invalid arguments")
	}

	mapVal := reflect.ValueOf(kv)
	if mapVal.Kind() == reflect.Ptr {
		mapVal = mapVal.Elem()
	}

	if mapVal.Kind() != reflect.Map {
		return errors.New("kv must a map")
	}

	mapLen := mapVal.Len()
	if mapLen < 1 {
		return errors.New("kv must not an empty map")
	}

	array := make([]interface{}, 2*mapLen+1)
	array[0] = key
	i := 1
	for _, item := range mapVal.MapKeys() {
		array[i] = item.Interface()
		i++
		array[i] = mapVal.MapIndex(item).Interface()
		i++
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("HMSET", array...)
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

// 执行0个参数的脚本
func DoLuaWith0Arg(dbIndex int, script string, resultPtr interface{}) error {
	if dbIndex < 0 || dbIndex > 15 {
		return errors.New("dbIndex must in [0,15]")
	}

	if resultPtr != nil {
		resultValRef := reflect.ValueOf(resultPtr)
		if resultValRef.Kind() != reflect.Ptr {
			return errors.New("resultPtr must be a pointer")
		}
	}

	conn := getConn0(uint8(dbIndex))
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	reply, err := redis.NewScript(0, script).Do(conn)
	if err != nil {
		return err
	} else if resultPtr == nil {
		return nil
	}

	return parse(reply, resultPtr)
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

// 事务提交redis命令
func MultiExec(key string, commands [][]interface{}) error {
	if len(commands) < 1 {
		return nil
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	for i, cmd := range commands {
		if len(cmd) < 2 {
			return errors.New(fmt.Sprintf("invalid command at position %d", i))
		}

		key := cmd[0]
		cmdName, ok := key.(string)
		if !ok {
			return errors.New(fmt.Sprintf("command name not string at position %d", i))
		}

		args := cmd[1:]

		if err := conn.Send(cmdName, args...); err != nil {
			return err
		}
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
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

// 获取所有key val
func HashGetAll(key string, mapPtr interface{}) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid key")
	}

	if mapPtr == nil {
		return errors.New("mapPtr must not be nil")
	}

	mapValRef := reflect.ValueOf(mapPtr)
	if mapValRef.Kind() != reflect.Ptr || mapValRef.Elem().Kind() != reflect.Map {
		return errors.New("mapPtr must be a map pointer")
	}

	mapKeyKind := mapValRef.Elem().Type().Key().Kind()
	if mapKeyKind != reflect.String && mapKeyKind != reflect.Int && mapKeyKind != reflect.Int64 {
		return errors.New("map key must be string or int or int64")
	}

	mapValKind := mapValRef.Elem().Type().Elem().Kind()
	if mapValKind == reflect.Ptr {
		if mapValRef.Elem().Type().Elem().Elem().Kind() != reflect.Struct {
			return errors.New("map value must be string or int or int64 or float64 or struct or structPtr")
		}
	} else if mapValKind != reflect.String && mapValKind != reflect.Int && mapValKind != reflect.Int64 && mapValKind != reflect.Float64 && mapValKind != reflect.Struct {
		return errors.New("map value must be string or int or int64 or float64 or struct or structPtr")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("HGETALL", key)
	if err != nil {
		return err
	}

	tempSlice := make([]string, 0)
	err = parse(reply, &tempSlice)
	if err != nil {
		return err
	}

	for i := 0; i < len(tempSlice); i++ {
		if i%2 == 0 {
			var mapKey, mapVal reflect.Value
			metaKey, metaVal := tempSlice[i], tempSlice[i+1]

			if mapKeyKind == reflect.String {
				mapKey = reflect.ValueOf(metaKey)
			} else {
				if key, err := strconv.ParseInt(metaKey, 10, 64); err != nil {
					return err
				} else if mapKeyKind == reflect.Int {
					mapKey = reflect.ValueOf(int(key))
				} else {
					mapKey = reflect.ValueOf(key)
				}
			}

			if mapValKind == reflect.String {
				mapVal = reflect.ValueOf(metaVal)
			} else if mapValKind == reflect.Int || mapValKind == reflect.Int64 {
				if val, err := strconv.ParseInt(metaVal, 10, 64); err != nil {
					return err
				} else if mapValKind == reflect.Int {
					mapVal = reflect.ValueOf(int(val))
				} else {
					mapVal = reflect.ValueOf(val)
				}
			} else if mapValKind == reflect.Float64 {
				if val, err := strconv.ParseFloat(metaVal, 64); err != nil {
					return err
				} else {
					mapVal = reflect.ValueOf(val)
				}
			} else {
				// mapValKind == reflect.Ptr || mapValKind == reflect.Struct
				tp := mapValRef.Elem().Type().Elem()
				if mapValKind == reflect.Ptr {
					tp = tp.Elem()
				}

				ptr := reflect.New(tp)
				if err := utils.JsonToStruct(metaVal, ptr.Interface()); err != nil {
					return err
				}

				if mapKeyKind == reflect.Ptr {
					mapVal = ptr
				} else {
					mapVal = ptr.Elem()
				}
			}

			mapValRef.Elem().SetMapIndex(mapKey, mapVal)
		}
	}

	return nil
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

	keys := make([]interface{}, 0)
	fileds := reflect.ValueOf(hashField)
	if fileds.Kind() == reflect.Slice {
		for i := 0; i < fileds.Len(); i++ {
			keys = append(keys, fileds.Index(i).Interface())
		}
		if len(keys) < 1 {
			return nil
		}
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	if len(keys) > 0 {
		for _, field := range keys {
			conn.Send("HDEL", key, field)
		}

		_, err := conn.Do("")
		return err

	} else {
		_, err := conn.Do("HDEL", key, hashField)
		return err
	}
}

// 增减hash值
func HashIncr(key string, hashField interface{}, incrVal int, afterIncr *int) error {
	if utils.IsEmpty(key) || hashField == nil {
		return errors.New("invalid key or hashField")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("HINCRBY", key, hashField, incrVal)
	if err != nil {
		return err
	}

	if afterIncr == nil {
		return nil
	}

	if err := parse(reply, afterIncr); err != nil {
		return err
	}
	return nil
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
		if vi, e := redis.Int(reply, nil); e != nil {
			err = e
		} else {
			switch objVale.Kind() {
			case reflect.Int:
				valueInterface = vi
			case reflect.Int8:
				valueInterface = int8(vi)
			case reflect.Int16:
				valueInterface = int16(vi)
			case reflect.Int32:
				valueInterface = int32(vi)
			case reflect.Uint:
				valueInterface = uint(vi)
			case reflect.Uint8:
				valueInterface = uint8(vi)
			case reflect.Uint16:
				valueInterface = uint16(vi)
			default:
				valueInterface = uint32(vi)
			}
		}
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
