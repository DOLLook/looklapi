package loggers

import (
	"fmt"
	"log"
	"looklapi/common/appcontext"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"looklapi/errs"
	"reflect"
	"strings"
)

// 控制台日志
type consoleLogger struct {
	logLevel logLevel // 日志等级
}

func init() {
	var logger = &consoleLogger{logLevel: level(appConfig.AppConfig.Logger.InitLevel)}
	logger.setLogger()
	logger.Subscribe()
}

func (logger *consoleLogger) name() string {
	return "console"
}

func (logger *consoleLogger) setLogger() {
	setLogger(logger)
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (logger *consoleLogger) OnApplicationEvent(event interface{}) {
	if event, ok := event.(*ConfigLog); ok {
		logger.logLevel = logLevel(event.LogLevel)
	}
}

// register to the application event publisher
// @eventType the event type which the observer interested in
func (logger *consoleLogger) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(logger, reflect.TypeOf(&ConfigLog{}))
}

// 调试日志
func (logger *consoleLogger) Debug(msg string) {
	if logger.logLevel < _DEBUG {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	//methodName, fullFileName, _, lineNum := getTrace()
	//log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_DEBUG), methodName, msg, fullFileName, lineNum))

	methodName, _, _, _ := getTrace()
	methodNameSplit := strings.Split(methodName, "/")
	mn := "[" + methodNameSplit[len(methodNameSplit)-1] + "]"
	log.Println(fmt.Sprintf("%v %s %s", levelName(_DEBUG), mn, msg))
}

// 提示
func (logger *consoleLogger) Info(msg string) {
	if logger.logLevel < _INFO {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	//methodName, fullFileName, _, lineNum := getTrace()
	//log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_INFO)+" ", methodName, msg, fullFileName, lineNum))

	methodName, _, _, _ := getTrace()
	methodNameSplit := strings.Split(methodName, "/")
	mn := "[" + methodNameSplit[len(methodNameSplit)-1] + "]"
	log.Println(fmt.Sprintf("%v %s %s", levelName(_INFO)+" ", mn, msg))
}

// 警告
func (logger *consoleLogger) Warn(msg string) {
	if logger.logLevel < _WARN {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	//methodName, fullFileName, _, lineNum := getTrace()
	//log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_WARN)+" ", methodName, msg, fullFileName, lineNum))

	methodName, _, _, _ := getTrace()
	methodNameSplit := strings.Split(methodName, "/")
	mn := "[" + methodNameSplit[len(methodNameSplit)-1] + "]"
	log.Println(fmt.Sprintf("%v %s %s", levelName(_WARN)+" ", mn, msg))
}

// 错误日志
func (logger *consoleLogger) Error(err error) {
	if logger.logLevel < _ERROR {
		return
	}

	if err == nil {
		return
	}

	stackTrace := ""
	if berr, ok := err.(*errs.BllError); ok {
		var trace []string
		for _, stack := range berr.FormatStackTrace() {
			if stack.Invalid() {
				trace = append(trace, "\n\t"+stack.Method())
			} else {
				trace = append(trace, fmt.Sprintf("\n\t%s\n\t%s:%d", stack.Method(), stack.File(), stack.Line()))
			}
		}
		stackTrace = strings.Join(trace, "\n")

		log.Println(fmt.Sprintf("%v %s%s", levelName(_ERROR), err.Error(), stackTrace))
	} else {
		methodName, fullFileName, _, lineNum := getTrace()
		methodNameSplit := strings.Split(methodName, "/")
		mn := "[" + methodNameSplit[len(methodNameSplit)-1] + "]"
		log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_ERROR), mn, err.Error(), fullFileName, lineNum))
	}

}
