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
	"strconv"
	"strings"
)

// mongo日志
type mongoLogger struct {
	logLevel logLevel // 日志等级
}

func init() {
	var logger = &mongoLogger{logLevel: level(config.AppConfig.Logger.InitLevel)}
	if config.AppConfig.Logger.DefaultLogger != logger.name() {
		return
	}
	if !mongoutils.ClientIsValid() {
		panic(fmt.Sprintf("mongo client is not valid, please check the config"))
	}

	logger.setLogger()
	logger.Subscribe()
}

func (logger *mongoLogger) name() string {
	return "mongo"
}

func (logger *mongoLogger) setLogger() {
	setLogger(logger)
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (logger *mongoLogger) OnApplicationEvent(event interface{}) {
	if event, ok := event.(*ConfigLog); ok {
		logger.logLevel = logLevel(event.LogLevel)
	}
}

// register to the application event publisher
// @eventType the event type which the observer interested in
func (logger *mongoLogger) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(logger, reflect.TypeOf(&ConfigLog{}))
}

// 调试日志
func (logger *mongoLogger) Debug(msg string) {
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
func (logger *mongoLogger) Info(msg string) {
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
func (logger *mongoLogger) Warn(msg string) {
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
func (logger *mongoLogger) Error(err error) {
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
		for _, stack := range berr.FormatStackTrace() {
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
	} else if serr, ok := err.(stackTracer); ok {
		st := serr.StackTrace()
		stack := fmt.Sprintf("%+v", st)
		lines := make([]string, 0)
		prefix := ""
		for _, line := range strings.Split(stack, "\n") {
			if len(strings.TrimSpace(line)) < 1 {
				continue
			}

			fileLine := false
			splits := strings.Split(line, ":")
			if len(splits) > 1 {
				if _, err := strconv.Atoi(splits[len(splits)-1]); err == nil {
					fileLine = true
				}
			}

			if fileLine {
				l := "\t" + prefix + strings.Split(line, prefix)[1]
				lines = append(lines, l)
			} else {
				splits := strings.Split(line, "/")
				if len(splits) > 1 {
					prefix = splits[0] + "/"
				} else {
					prefix = strings.Split(line, ".")[0] + "/"
				}

				lines = append(lines, line)
			}
		}

		splitsL0 := strings.Split(lines[1], "/")
		log.ClassName = strings.Split(splitsL0[len(splitsL0)-1], ":")[0]
		log.Stacktrace = strings.Join(lines, "\n")
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
