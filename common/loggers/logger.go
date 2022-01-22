package loggers

import (
	"looklapi/common/appcontext"
	"looklapi/common/redisutils"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"reflect"
	"runtime"
	"strings"
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

	// 名称
	name() string

	setLogger()
}

type logManager struct {
}

// recieved app event and process.
// for event publish well, the developers must deal with the panic by their self
func (manager *logManager) OnApplicationEvent(event interface{}) {
	logConfig := &ConfigLog{}
	if err := redisutils.Get(redisutils.CONFIG_LOG, logConfig); err != nil {
		GetLogger().Error(err)
	} else {
		appcontext.GetAppEventPublisher().PublishEvent(logConfig)
	}
}

// regiser to the application event publisher
func (manager *logManager) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(manager, reflect.TypeOf(appcontext.AppEventBeanInjected(0)))
}

func init() {
	manager := &logManager{}
	manager.Subscribe()
}

type ConfigLog struct {
	LogLevel int8
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
var _buildinLogger Logger // 内置logger

func setLogger(logger Logger) {
	if logger == nil {
		return
	}
	if utils.ArrayOrSliceContains(_loggers, logger) {
		return
	}

	_loggers = append(_loggers, logger)

	if logger.name() == appConfig.AppConfig.Logger.Default {
		_defaultLogger = logger
	}

	if logger.name() == "buildin" {
		_buildinLogger = logger
	}
}

// 获取logger
func GetLogger() Logger {
	return _defaultLogger
}

// 获取内置looger
func GetBuildinLogger() Logger {
	return _buildinLogger
}

//func getTrace(){
//stackStr := string(debug.Stack())
//stackSlice := strings.Split(stackStr, "\n")
//if level == _ERROR || level == _FATAL {
//	var temp []string
//	temp = append(temp, stackSlice[0])
//	temp = append(temp, stackSlice[7:]...)
//	log.Stacktrace = strings.Join(temp, "\n")
//}
//
//if routineId, err := strconv.Atoi(strings.Split(stackSlice[0], " ")[1]); err == nil {
//	log.ThreadId = int32(routineId)
//}
//}

func getTrace() (methodName string, fullFileName string, fileName string, lineNum int) {
	methodName, fullFileName, fileName = "", "", ""
	lineNum = 0
	pc, fullFileName, lineNum, ok := runtime.Caller(2)
	if ok {
		methodName = runtime.FuncForPC(pc).Name()
	}
	fullFileName = strings.TrimSpace(fullFileName)
	if len(fullFileName) > 0 {
		indexNum := strings.Index(fullFileName, "/src/")
		fullFileName = fullFileName[indexNum+4:]
		temp := strings.Split(fullFileName, "/")
		fileName = temp[len(temp)-1]
	}

	return
}

func levelName(level logLevel) string {
	switch level {
	case _DEBUG:
		return "DEBUG"
	case _INFO:
		return "INFO"
	case _WARN:
		return "WARN"
	case _ERROR:
		return "ERROR"
	case _FATAL:
		return "FATAL"
	default:
		return ""
	}
}
