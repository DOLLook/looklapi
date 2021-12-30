package loggers

import (
	"fmt"
	"log"
	"micro-webapi/common/appcontext"
	"micro-webapi/common/utils"
	"micro-webapi/errs"
	"micro-webapi/model/modelimpl"
	"reflect"
	"strings"
)

type buildinLogger struct {
	logLevel logLevel // 日志等级
}

func init() {
	//var logger interface{} = &buildinLogger{}
	var logger = &buildinLogger{logLevel: _INFO}
	logger.setLogger()
	logger.Subscribe()
}

func (logger *buildinLogger) name() string {
	return "buildin"
}

func (logger *buildinLogger) setLogger() {
	setLogger(logger)
}

// recieved app event and process.
// for event publish well, the developers must deal with the panic by their self
func (logger *buildinLogger) OnApplicationEvent(event interface{}) {
	if event, ok := event.(*modelimpl.ConfigLog); ok {
		logger.logLevel = logLevel(event.LogLevel)
	}
}

// regiser to the application event publisher
// @eventType the event type which the observer intrested in
func (logger *buildinLogger) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(logger, reflect.TypeOf(&modelimpl.ConfigLog{}))
}

// 调试日志
func (logger *buildinLogger) Debug(msg string) {
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
func (logger *buildinLogger) Info(msg string) {
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
func (logger *buildinLogger) Warn(msg string) {
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
func (logger *buildinLogger) Error(err error) {
	if logger.logLevel < _ERROR {
		return
	}

	if err == nil {
		return
	}

	stackTrace := ""
	if berr, ok := err.(*errs.BllError); ok {
		var trace []string
		for _, stack := range berr.StackTrace() {
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
