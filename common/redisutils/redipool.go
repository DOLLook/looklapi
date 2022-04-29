package redisutils

import (
	"github.com/garyburd/redigo/redis"
	"looklapi/common/utils"
	"looklapi/config"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var address = config.AppConfig.Redis.Host + ":" + config.AppConfig.Redis.Port
var pwdOption = redis.DialPassword(config.AppConfig.Redis.Password)
var redisPool = &sync.Map{}

const (
	_NETWORK    = "tcp"
	_MAXDBINDEX = 16
)

// 获取数据库连接
func getConn(key string) redis.Conn {
	strSlice := strings.Split(key, "_")
	if len(strSlice) < 2 {
		return getConn0(0)
	}

	reg := regexp.MustCompile("\\d+")
	match := reg.FindString(strSlice[0])
	dbIndex := 0
	if !utils.IsEmpty(match) {
		if dbIndex, err := strconv.Atoi(match); err == nil && dbIndex < _MAXDBINDEX {
			return getConn0(uint8(dbIndex))
		}
	}

	return getConn0(uint8(dbIndex))
}

// 获取数据库连接
func getConn0(db uint8) redis.Conn {
	pool, ok := redisPool.Load(db)
	if ok && pool != nil {
		return pool.(*redis.Pool).Get()
	}

	pool0 := newPool(db)
	redisPool.Store(db, pool0)

	return pool0.Get()
}

func newPool(db uint8) *redis.Pool {
	return &redis.Pool{ //实例化一个连接池
		MaxIdle:     16,  //最初的连接数量
		MaxActive:   500, //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: 300, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Wait:        true,
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			return redis.Dial(_NETWORK, address, pwdOption, redis.DialDatabase(int(db)))
		},
	}
}
