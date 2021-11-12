package loggers

import (
	"go-webapi-fw/common/redisutils"
	"go-webapi-fw/common/utils"
	appConfig "go-webapi-fw/config"
	"go-webapi-fw/model/modelimpl"
)

type Logger interface {
	// 调试日志
	Debug(msg string)

	// 提示
	Info(msg string)

	// 警告
	Warn(msg string)

	// 错误日志
	Error(err error)

	// 通知日志等级刷新
	notifyLoglevel(level logLevel)

	// 名称
	name() string
}

// 日志等级别名
type logLevel = byte

// 日志等级枚举
const (
	_OFF logLevel = iota
	_FATAL
	_ERROR
	_WARN
	_INFO
	_DEBUG
	_ALL
)

var _level = _OFF
var _loggers []Logger     // logger容器
var _defaultLogger Logger // 默认logger

func setLogger(logger Logger) {
	if logger == nil {
		return
	}
	if utils.ArrayOrSliceContains(_loggers, logger) {
		return
	}

	appendLogLevel(logger)

	_loggers = append(_loggers, logger)

	if logger.name() == appConfig.AppConfig.Logger.Default {
		_defaultLogger = logger
	}
}

// 更新日志等级
func RefreshLogLevel(level int32) {
	for _, logger := range _loggers {
		logger.notifyLoglevel(logLevel(level))
	}
}

// 获取logger
func GetLogger() Logger {
	return _defaultLogger
}

func appendLogLevel(logger Logger) {
	if _level > _OFF {
		logger.notifyLoglevel(_level)
		return
	}

	configLog := &modelimpl.ConfigLog{}
	redisutils.Get(redisutils.CONFIG_LOG, configLog)
	_level = logLevel(configLog.LogLevel)
	logger.notifyLoglevel(_level)
}
