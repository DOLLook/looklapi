package redisutils

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

// 加锁执行, 执行完成自动释放锁
func LockAction(action func(), lockName string, timeoutSecs int32) {
	if ok, timeStamp := TryLock(lockName, timeoutSecs); ok {
		defer UnLock(lockName, timeStamp)
		action()
	}
}

// 加锁
func TryLock(lockName string, timeoutSecs int32) (bool, int64) {
	begin := time.Now().UnixNano()
	for {
		result, timeStamp := lock(lockName, 30)
		if result {
			return true, timeStamp
		}

		time.Sleep(1 * time.Millisecond)
		if time.Now().UnixNano()-begin >= int64(timeoutSecs*1000*1000*1000) {
			result = false
			break
		}
	}

	return false, 0
}

// 加锁
func lock(lockName string, holdSecs int32) (bool, int64) {
	lockName = getLockName(lockName)

	scriptStr := `if redis.call('EXISTS',KEYS[1])==0 then
				redis.call('SET',KEYS[1],ARGV[1])
				return redis.call('EXPIRE',KEYS[1],ARGV[2])
				else return 0
				end`

	script := redis.NewScript(1, scriptStr)
	conn := getConn0(0)
	timeStamp := time.Now().UnixNano() / 1000000
	reply, err := script.Do(conn, lockName, timeStamp, holdSecs)

	if err != nil {
		panic(err)
	}

	var result bool
	parse(reply, &result)
	return result, timeStamp
}

// 解锁
func UnLock(lockName string, timeStamp int64) {
	lockName = getLockName(lockName)

	scriptStr := `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) else return 0 end`

	script := redis.NewScript(1, scriptStr)
	conn := getConn0(0)
	_, err := script.Do(conn, lockName, timeStamp)
	if err != nil {
		panic(err)
	}
}

// 获取真实锁名称
func getLockName(lockName string) string {
	return "locker_" + lockName
}
