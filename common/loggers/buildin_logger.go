package loggers

import (
	"fmt"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/errs"
	"log"
	"strings"
)

type buildinLogger struct {
	logLevel logLevel // 日志等级
}

func init() {
	var logger interface{} = &buildinLogger{}
	setLogger(logger.(Logger))
}

func (logger *buildinLogger) name() string {
	return "buildin"
}

func (logger *buildinLogger) notifyLoglevel(level logLevel) {
	logger.logLevel = level
}

// 调试日志
func (logger *buildinLogger) Debug(msg string) {
	if logger.logLevel < _DEBUG {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	methodName, fullFileName, _, lineNum := getTrace()
	log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_DEBUG), methodName, msg, fullFileName, lineNum))
}

// 提示
func (logger *buildinLogger) Info(msg string) {
	if logger.logLevel < _INFO {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	methodName, fullFileName, _, lineNum := getTrace()
	log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_INFO)+" ", methodName, msg, fullFileName, lineNum))
}

// 警告
func (logger *buildinLogger) Warn(msg string) {
	if logger.logLevel < _WARN {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	methodName, fullFileName, _, lineNum := getTrace()
	log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_WARN)+" ", methodName, msg, fullFileName, lineNum))
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
	if err, ok := err.(*errs.BllError); ok {
		var trace []string
		for _, stack := range err.StackTrace() {
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
		log.Println(fmt.Sprintf("%v %s %s\n\t%s:%d", levelName(_ERROR), methodName, err.Error(), fullFileName, lineNum))
	}

}
