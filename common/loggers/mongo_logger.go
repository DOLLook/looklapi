package loggers

import (
	"fmt"
	"looklapi/common/appcontext"
	"looklapi/common/mongoutils"
	"looklapi/common/utils"
	"looklapi/config"
	"looklapi/errs"
	"looklapi/model/mongo"
	"reflect"
	"strings"
)

type mongoLoger struct {
	logLevel logLevel // 日志等级
}

func init() {
	//var logger interface{} = &mongoLoger{}
	var logger = &mongoLoger{logLevel: _INFO}
	logger.setLogger()
	logger.Subscribe()
}

func (logger *mongoLoger) name() string {
	return "mongo"
}

func (logger *mongoLoger) setLogger() {
	setLogger(logger)
}

// recieved app event and process.
// for event publish well, the developers must deal with the panic by their self
func (logger *mongoLoger) OnApplicationEvent(event interface{}) {
	if event, ok := event.(*ConfigLog); ok {
		logger.logLevel = logLevel(event.LogLevel)
	}
}

// regiser to the application event publisher
// @eventType the event type which the observer intrested in
func (logger *mongoLoger) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(logger, reflect.TypeOf(&ConfigLog{}))
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

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(log.Content)
				fmt.Println(log.Stacktrace)
				if tr, ok := err.(error); ok {
					fmt.Println(tr.Error())
				} else if msg, ok := err.(string); ok {
					fmt.Println(msg)
				}
			}
		}()

		if _, err := mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log); err != nil {
			fmt.Println(log.Content)
			fmt.Println(log.Stacktrace)
			fmt.Println(err.Error())
		}
	}()
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

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(log.Content)
				fmt.Println(log.Stacktrace)
				if tr, ok := err.(error); ok {
					fmt.Println(tr.Error())
				} else if msg, ok := err.(string); ok {
					fmt.Println(msg)
				}
			}
		}()

		if _, err := mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log); err != nil {
			fmt.Println(log.Content)
			fmt.Println(log.Stacktrace)
			fmt.Println(err.Error())
		}
	}()
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

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(log.Content)
				fmt.Println(log.Stacktrace)
				if tr, ok := err.(error); ok {
					fmt.Println(tr.Error())
				} else if msg, ok := err.(string); ok {
					fmt.Println(msg)
				}
			}
		}()

		if _, err := mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log); err != nil {
			fmt.Println(log.Content)
			fmt.Println(log.Stacktrace)
			fmt.Println(err.Error())
		}
	}()
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

	if berr, ok := err.(*errs.BllError); ok {
		var trace []string
		for _, stack := range berr.StackTrace() {
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

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(log.Content)
				fmt.Println(log.Stacktrace)
				if tr, ok := err.(error); ok {
					fmt.Println(tr.Error())
				} else if msg, ok := err.(string); ok {
					fmt.Println(msg)
				}
			}
		}()

		if _, err := mongoutils.GetCollection(log.TbCollName()).InsertOne(nil, log); err != nil {
			fmt.Println(log.Content)
			fmt.Println(log.Stacktrace)
			fmt.Println(err.Error())
		}
	}()
}
