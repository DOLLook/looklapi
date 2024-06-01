package redisutils

import (
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

// 加锁执行, 执行完成自动释放锁
func LockAction(action func() error, lockName string, timeoutSecs int32) (bool, error) {
	if ok, timeStamp, err := TryLock(lockName, timeoutSecs); ok {
		defer UnLock(lockName, timeStamp)
		return ok, action()
	} else {
		return false, err
	}
}

// 加锁
func TryLock(lockName string, timeoutSecs int32) (bool, int64, error) {
	begin := time.Now().UnixNano()
	for {
		result, timeStamp, err := lock(lockName, 30)
		if err != nil {
			return false, 0, err
		}
		if result {
			return true, timeStamp, nil
		}

		time.Sleep(1 * time.Millisecond)
		if time.Now().UnixNano()-begin >= int64(timeoutSecs)*1000*1000*1000 {
			result = false
			break
		}
	}

	return false, 0, nil
}

// 加锁
func lock(lockName string, holdSecs int32) (bool, int64, error) {
	//lockName = getLockName(lockName)
	//
	//scriptStr := `if redis.call('EXISTS',KEYS[1])==0 then
	//			redis.call('SET',KEYS[1],ARGV[1])
	//			return redis.call('EXPIRE',KEYS[1],ARGV[2])
	//			else return 0
	//			end`
	//
	//script := redis.NewScript(1, scriptStr)
	//conn := getConn0(0)
	//timeStamp := time.Now().UnixNano() / 1000000
	//reply, err := script.Do(conn, lockName, timeStamp, holdSecs)
	//
	//if err != nil {
	//	return false, 0, err
	//}
	//
	//var result bool
	//if err := parse(reply, &result); err != nil {
	//	return false, 0, err
	//}
	//return result, timeStamp, nil

	lockName = getLockName(lockName)
	conn := getConn0(0)
	if conn.Err() != nil {
		return false, 0, conn.Err()
	}
	defer conn.Close()

	timeStamp := time.Now().UnixNano() / 1000000
	reply, err := conn.Do("SET", lockName, timeStamp, "EX", holdSecs, "NX")

	if err != nil || reply == nil {
		return false, 0, err
	}

	if reply, ok := reply.(string); ok && strings.ToLower(reply) == "ok" {
		return true, timeStamp, nil
	} else {
		return false, 0, nil
	}
}

// 解锁
func UnLock(lockName string, timeStamp int64) error {
	lockName = getLockName(lockName)

	scriptStr := `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) else return 0 end`

	script := redis.NewScript(1, scriptStr)
	conn := getConn0(0)
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := script.Do(conn, lockName, timeStamp)
	return err
}

// 获取真实锁名称
func getLockName(lockName string) string {
	return "locker_" + lockName
}
