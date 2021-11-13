package loggers

import (
	"fmt"
	"go-webapi-fw/common/mongoutils"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/config"
	"go-webapi-fw/errs"
	"go-webapi-fw/model/mongo"
	"strings"
)

type mongoLoger struct {
	logLevel logLevel // 日志等级
}

func init() {
	var logger interface{} = &mongoLoger{}
	setLogger(logger.(Logger))
}

func (logger *mongoLoger) name() string {
	return "mongo"
}

func (logger *mongoLoger) notifyLoglevel(level logLevel) {
	logger.logLevel = level
}

// 调试日志
func (logger *mongoLoger) Debug(msg string) {
	if logger.logLevel < _DEBUG {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	methodName, fullFileName, fileName, lineNum := getTrace()
	log := mongo.NewMongoLog()
	log.Instance = config.AppConfig.Server.Name
	log.HostIp = utils.HostIp()
	log.Content = msg
	log.Level = int32(_DEBUG)

	log.ClassName = fileName
	log.Stacktrace = fmt.Sprintf("%s\n\t%s:%d", methodName, fullFileName, lineNum)

	go mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log)
}

// 提示
func (logger *mongoLoger) Info(msg string) {
	if logger.logLevel < _INFO {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := mongo.NewMongoLog()
	log.Instance = config.AppConfig.Server.Name
	log.HostIp = utils.HostIp()
	log.Content = msg
	log.Level = int32(_INFO)

	methodName, fullFileName, fileName, lineNum := getTrace()
	log.ClassName = fileName
	log.Stacktrace = fmt.Sprintf("%s\n\t%s:%d", methodName, fullFileName, lineNum)

	go mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log)
}

// 警告
func (logger *mongoLoger) Warn(msg string) {
	if logger.logLevel < _WARN {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := mongo.NewMongoLog()
	log.Instance = config.AppConfig.Server.Name
	log.HostIp = utils.HostIp()
	log.Content = msg
	log.Level = int32(_WARN)

	methodName, fullFileName, fileName, lineNum := getTrace()
	log.ClassName = fileName
	log.Stacktrace = fmt.Sprintf("%s\n\t%s:%d", methodName, fullFileName, lineNum)

	go mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log)
}

// 错误日志
func (logger *mongoLoger) Error(err error) {
	if logger.logLevel < _ERROR {
		return
	}

	if err == nil {
		return
	}

	log := mongo.NewMongoLog()
	log.Instance = config.AppConfig.Server.Name
	log.HostIp = utils.HostIp()
	log.Content = err.Error()
	log.Level = int32(_ERROR)

	if err, ok := err.(*errs.BllError); ok {
		var trace []string
		for _, stack := range err.StackTrace() {
			if stack.Invalid() {
				trace = append(trace, stack.Method()+"\n")
			} else {
				trace = append(trace, fmt.Sprintf("%s\n\t%s:%d", stack.Method(), stack.File(), stack.Line()))
			}

			if utils.IsEmpty(log.ClassName) {
				log.ClassName = stack.FileName()
			}
		}
		log.Stacktrace = strings.Join(trace, "\n")
	} else {
		methodName, fullFileName, fileName, lineNum := getTrace()
		log.ClassName = fileName
		log.Stacktrace = fmt.Sprintf("%s\n\t%s:%d", methodName, fullFileName, lineNum)
	}

	go mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log)
}
