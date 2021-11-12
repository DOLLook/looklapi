package redisutils

type redisKey = string

const (
	CONFIG_MANUAL_SERVICE redisKey = "config_manual_service" // 手动服务配置
	CONFIG_LOG            redisKey = "config_log"            // 日志配置
)
