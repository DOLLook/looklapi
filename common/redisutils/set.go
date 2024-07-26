package redisutils

import (
	"errors"
	"looklapi/common/utils"
)

// 添加值 返回成功添加的数量
func SetAdd(key string, members ...string) (int, error) {
	if utils.IsEmpty(key) || len(members) < 1 {
		return 0, errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	defer conn.Close()

	cmd := make([]any, 0)
	cmd = append(cmd, key)
	for _, m := range members {
		cmd = append(cmd, m)
	}
	reply, err := conn.Do("SADD", cmd...)
	if err != nil {
		return 0, err
	}

	result := 0
	if err := parse(reply, &result); err != nil {
		return 0, err
	} else {
		return result, nil
	}
}

// 计数
func SetCount(key string) (int, error) {
	if utils.IsEmpty(key) {
		return 0, errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("SCARD", key)
	if err != nil {
		return 0, err
	}

	result := 0
	if err := parse(reply, &result); err != nil {
		return 0, err
	} else {
		return result, nil
	}
}

// 测试member是否存在
func SetExist(key string, member string) (bool, error) {
	if utils.IsEmpty(key) || utils.IsEmpty(member) {
		return false, errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return false, conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("SISMEMBER", key, member)
	if err != nil {
		return false, err
	}

	result := false
	if err := parse(reply, &result); err != nil {
		return false, err
	} else {
		return result, nil
	}
}

// 测试member是否存在
func SetAllExist(key string, members ...string) (bool, error) {
	if utils.IsEmpty(key) || len(members) < 1 {
		return false, errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return false, conn.Err()
	}
	defer conn.Close()

	cmd := make([]any, 0)
	cmd = append(cmd, key)
	for _, m := range members {
		cmd = append(cmd, m)
	}
	reply, err := conn.Do("SMISMEMBER", cmd...)
	if err != nil {
		return false, err
	}

	result := false
	if err := parse(reply, &result); err != nil {
		return false, err
	} else {
		return result, nil
	}
}

// 获取SET所有成员
func SetMembers(key string, members *[]string) error {
	if utils.IsEmpty(key) || members == nil {
		return errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	reply, err := conn.Do("SMEMBERS", key)
	if err != nil {
		return err
	}

	if err := parse(reply, members); err != nil {
		return err
	} else {
		return nil
	}
}

// 删除成员 返回删除的数量
func SetRemove(key string, members ...string) (int, error) {
	if utils.IsEmpty(key) || len(members) < 1 {
		return 0, errors.New("invalid arguments")
	}

	conn := getConn(key)
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	defer conn.Close()

	cmd := make([]any, 0)
	cmd = append(cmd, key)
	for _, m := range members {
		cmd = append(cmd, m)
	}
	reply, err := conn.Do("SREM", cmd...)
	if err != nil {
		return 0, err
	}

	result := 0
	if err := parse(reply, &result); err != nil {
		return 0, err
	} else {
		return result, nil
	}
}
