package redisutils

import (
	"errors"
	"looklapi/common/utils"
	"reflect"
	"strconv"
)

// 添加值
func ZAdd(key string, member string, score int64) error {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("ZADD", key, score, member)

	return err
}

// 查询值
func ZScore(key string, member string, valPtr interface{}) error {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return errors.New("invalid arguments")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("ZSCORE", key, member)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 删除成员
func ZRemove(key string, member string) error {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("ZREM", key, member)

	return err
}

// 查询[minScore,maxScore]区间的成员数量
func ZCount(key string, minScore int64, maxScore int64, valPtr interface{}) error {
	if utils.IsEmpty(key) || minScore > maxScore {
		return errors.New("invalid arguments")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("ZCOUNT", key, minScore, maxScore)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 查询成员以score在zset中从小到大的排序号 当member不在zset中时返回err.Nil
func ZRank(key string, member string, valPtr interface{}) error {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return errors.New("invalid arguments")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("ZRANK", key, member)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 查询成员以score在zset中从大到小的排序号 当member不在zset中时返回err.Nil
func ZRevRank(key string, member string, valPtr interface{}) error {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return errors.New("invalid arguments")
	}

	if valPtr == nil {
		return errors.New("valPtr must not be nil")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("ZREVRANK", key, member)
	if err != nil {
		return err
	}

	return parse(reply, valPtr)
}

// 移除[minScore,maxScore]区间的成员
func ZRemByScore(key string, minScore int64, maxScore int64) error {
	if utils.IsEmpty(key) || minScore > maxScore {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYSCORE", key, minScore, maxScore)

	return err
}

// 移除[start,stop]位置区间的成员
// start stop 为成员以score在zset中从小到大的位置索引
// 序号可以为负数，如-1表示最后一个成员，-2表示倒数第二个成员
func ZRemByRank(key string, start int, stop int) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYRANK", key, start, stop)

	return err
}

// 获取[start,stop]位置区间的成员数据
// start stop 为成员以score在zset中从小到大的位置索引
// 序号可以为负数，如-1表示最后一个成员，-2表示倒数第二个成员
// 当withScores=true时，返回值接收必须为map，key为成员，val为score。当为false时，返回值接收必须是slice
func ZRange(key string, start int, stop int, sliceOrMapPtr interface{}, withScores bool) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	if sliceOrMapPtr == nil {
		return errors.New("sliceOrMapPtr must not be nil")
	}

	sliceOrMapVal := reflect.ValueOf(sliceOrMapPtr)
	if sliceOrMapVal.Kind() != reflect.Ptr {
		return errors.New("sliceOrMapPtr must be a sliceOrMap pointer")
	}

	if withScores {
		if sliceOrMapVal.Elem().Kind() != reflect.Map {
			return errors.New("sliceOrMapPtr must be a map pointer")
		}
	} else {
		if sliceOrMapVal.Elem().Kind() != reflect.Slice {
			return errors.New("sliceOrMapPtr must be a slice pointer")
		}
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	var reply interface{}
	var err error
	if withScores {
		reply, err = conn.Do("ZRANGE", key, start, stop, "WITHSCORES")
	} else {
		reply, err = conn.Do("ZRANGE", key, start, stop)
	}

	if err != nil {
		return err
	}

	if !withScores {
		return parse(reply, sliceOrMapPtr)
	}

	tempSlice := make([]string, 0)
	err = parse(reply, &tempSlice)
	if err != nil {
		return err
	}

	for i := 0; i < len(tempSlice); i++ {
		if i%2 == 0 {
			if val, err := strconv.Atoi(tempSlice[i+1]); err != nil {
				return err
			} else {
				sliceOrMapVal.Elem().SetMapIndex(reflect.ValueOf(tempSlice[i]), reflect.ValueOf(val))
			}
		}
	}

	return nil
}

// 获取[start,stop]位置区间的成员数据
// start stop 为成员以score在zset中从大到小的位置索引
// 序号可以为负数，如-1表示最后一个成员，-2表示倒数第二个成员
// 当withScores=true时，返回值接收必须为map，key为成员，val为score。当为false时，返回值接收必须是slice
func ZRevRange(key string, start int, stop int, sliceOrMapPtr interface{}, withScores bool) error {
	if utils.IsEmpty(key) {
		return errors.New("invalid arguments")
	}

	if sliceOrMapPtr == nil {
		return errors.New("sliceOrMapPtr must not be nil")
	}

	sliceOrMapVal := reflect.ValueOf(sliceOrMapPtr)
	if sliceOrMapVal.Kind() != reflect.Ptr {
		return errors.New("sliceOrMapPtr must be a sliceOrMap pointer")
	}

	if withScores {
		if sliceOrMapVal.Elem().Kind() != reflect.Map {
			return errors.New("sliceOrMapPtr must be a map pointer")
		}
	} else {
		if sliceOrMapVal.Elem().Kind() != reflect.Slice {
			return errors.New("sliceOrMapPtr must be a slice pointer")
		}
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	var reply interface{}
	var err error
	if withScores {
		reply, err = conn.Do("ZREVRANGE", key, start, stop, "WITHSCORES")
	} else {
		reply, err = conn.Do("ZREVRANGE", key, start, stop)
	}

	if err != nil {
		return err
	}

	if !withScores {
		return parse(reply, sliceOrMapPtr)
	}

	tempSlice := make([]string, 0)
	err = parse(reply, &tempSlice)
	if err != nil {
		return err
	}

	for i := 0; i < len(tempSlice); i++ {
		if i%2 == 0 {
			if val, err := strconv.Atoi(tempSlice[i+1]); err != nil {
				return err
			} else {
				sliceOrMapVal.Elem().SetMapIndex(reflect.ValueOf(tempSlice[i]), reflect.ValueOf(val))
			}
		}
	}

	return nil
}

// 获取[minScore,maxScore]区间的从小到大排列的成员数据
// 当withScores=true时，返回值接收必须为map，key为成员，val为score。当为false时，返回值接收必须是slice
func ZRangeByScore(key string, minScore int64, maxScore int64, sliceOrMapPtr interface{}, withScores bool) error {
	if utils.IsEmpty(key) || minScore > maxScore {
		return errors.New("invalid arguments")
	}

	if sliceOrMapPtr == nil {
		return errors.New("sliceOrMapPtr must not be nil")
	}

	sliceOrMapVal := reflect.ValueOf(sliceOrMapPtr)
	if sliceOrMapVal.Kind() != reflect.Ptr {
		return errors.New("sliceOrMapPtr must be a sliceOrMap pointer")
	}

	if withScores {
		if sliceOrMapVal.Elem().Kind() != reflect.Map {
			return errors.New("sliceOrMapPtr must be a map pointer")
		}
	} else {
		if sliceOrMapVal.Elem().Kind() != reflect.Slice {
			return errors.New("sliceOrMapPtr must be a slice pointer")
		}
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	var reply interface{}
	var err error
	if withScores {
		reply, err = conn.Do("ZRANGEBYSCORE", key, minScore, maxScore, "WITHSCORES")
	} else {
		reply, err = conn.Do("ZRANGEBYSCORE", key, minScore, maxScore)
	}

	if err != nil {
		return err
	}

	if !withScores {
		return parse(reply, sliceOrMapPtr)
	}

	tempSlice := make([]string, 0)
	err = parse(reply, &tempSlice)
	if err != nil {
		return err
	}

	for i := 0; i < len(tempSlice); i++ {
		if i%2 == 0 {
			if val, err := strconv.Atoi(tempSlice[i+1]); err != nil {
				return err
			} else {
				sliceOrMapVal.Elem().SetMapIndex(reflect.ValueOf(tempSlice[i]), reflect.ValueOf(val))
			}
		}
	}

	return nil
}

// 获取[maxScore,minScore]区间的从大到小排列的成员数据
// 当withScores=true时，返回值接收必须为map，key为成员，val为score。当为false时，返回值接收必须是slice
func ZRevRangeByScore(key string, maxScore int64, minScore int64, sliceOrMapPtr interface{}, withScores bool) error {
	if utils.IsEmpty(key) || minScore > maxScore {
		return errors.New("invalid arguments")
	}

	if sliceOrMapPtr == nil {
		return errors.New("sliceOrMapPtr must not be nil")
	}

	sliceOrMapVal := reflect.ValueOf(sliceOrMapPtr)
	if sliceOrMapVal.Kind() != reflect.Ptr {
		return errors.New("sliceOrMapPtr must be a sliceOrMap pointer")
	}

	if withScores {
		if sliceOrMapVal.Elem().Kind() != reflect.Map {
			return errors.New("sliceOrMapPtr must be a map pointer")
		}
	} else {
		if sliceOrMapVal.Elem().Kind() != reflect.Slice {
			return errors.New("sliceOrMapPtr must be a slice pointer")
		}
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	var reply interface{}
	var err error
	if withScores {
		reply, err = conn.Do("ZREVRANGEBYSCORE", key, maxScore, minScore, "WITHSCORES")
	} else {
		reply, err = conn.Do("ZREVRANGEBYSCORE", key, maxScore, minScore)
	}

	if err != nil {
		return err
	}

	if !withScores {
		return parse(reply, sliceOrMapPtr)
	}

	tempSlice := make([]string, 0)
	err = parse(reply, &tempSlice)
	if err != nil {
		return err
	}

	for i := 0; i < len(tempSlice); i++ {
		if i%2 == 0 {
			if val, err := strconv.Atoi(tempSlice[i+1]); err != nil {
				return err
			} else {
				sliceOrMapVal.Elem().SetMapIndex(reflect.ValueOf(tempSlice[i]), reflect.ValueOf(val))
			}
		}
	}

	return nil
}
