package loggers

import (
	"looklapi/common/appcontext"
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

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (manager *logManager) OnApplicationEvent(event interface{}) {
	// todo may be reset log level from load remote config when AppEventBeanInjected
	//logConfig := &ConfigLog{}
	//appcontext.GetAppEventPublisher().PublishEvent(logConfig)
}

// register to the application event publisher
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
var _consoleLogger Logger // 内置logger

func setLogger(logger Logger) {
	if logger == nil {
		return
	}
	if utils.ArrayOrSliceContains(_loggers, logger) {
		return
	}

	_loggers = append(_loggers, logger)

	if logger.name() == appConfig.AppConfig.Logger.DefaultLogger {
		_defaultLogger = logger
	}

	if logger.name() == "console" {
		_consoleLogger = logger
	}
}

// 获取logger
func GetLogger() Logger {
	return _defaultLogger
}

// 获取内置looker
func GetConsoleLogger() Logger {
	return _consoleLogger
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
	case _ALL:
		return "ALL"
	default:
		return ""
	}
}

func level(level string) logLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return _DEBUG
	case "INFO":
		return _INFO
	case "WARN":
		return _WARN
	case "ERROR":
		return _ERROR
	case "FATAL":
		return _FATAL
	case "ALL":
		return _ALL
	case "OFF":
		return _OFF
	default:
		return _INFO
	}
}
