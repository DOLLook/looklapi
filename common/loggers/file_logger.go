package loggers

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"looklapi/common/appcontext"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

// 文件日志
type fileLogger struct {
	logLevel   logLevel // 日志等级
	zeroLogger *zerolog.Logger
}

func init() {
	var logger = &fileLogger{logLevel: level(appConfig.AppConfig.Logger.InitLevel)}
	if appConfig.AppConfig.Logger.DefaultLogger != logger.name() {
		return
	}

	logger.setLogger()
	logger.Subscribe()
}

func (logger *fileLogger) name() string {
	return "file"
}

func (logger *fileLogger) setLogger() {
	if myLogger, err := initLogger(); err != nil {
		panic(err)
	} else {
		logger.zeroLogger = myLogger
	}

	setLogger(logger)
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (logger *fileLogger) OnApplicationEvent(event interface{}) {
	if event, ok := event.(*ConfigLog); ok {
		logger.logLevel = logLevel(event.LogLevel)
	}
}

// register to the application event publisher
// @eventType the event type which the observer interested in
func (logger *fileLogger) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(logger, reflect.TypeOf(&ConfigLog{}))
}

// 调试日志
func (logger *fileLogger) Debug(msg string) {
	if logger.logLevel < _DEBUG {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}
	logger.zeroLogger.Debug().Msg(msg)
}

// 提示
func (logger *fileLogger) Info(msg string) {
	if logger.logLevel < _INFO {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}
	logger.zeroLogger.Info().Msg(msg)
}

// 警告
func (logger *fileLogger) Warn(msg string) {
	if logger.logLevel < _WARN {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}
	logger.zeroLogger.Warn().Msg(msg)
}

// 错误日志
func (logger *fileLogger) Error(err error) {
	if logger.logLevel < _ERROR {
		return
	}

	if err == nil {
		return
	}
	logger.zeroLogger.Error().Stack().Err(err).Msg("")
}

func initLogger() (*zerolog.Logger, error) {
	zerolog.TimeFieldFormat = time.DateTime
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	parent := zerolog.New(os.Stderr).With().Timestamp().CallerWithSkipFrameCount(3).Logger()
	zConsole := parent.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime, NoColor: true})

	fileName := "log/app.log"
	dir := filepath.Dir(fileName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// 配置文件日志输出
	fileLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     10, // days
		LocalTime:  true,
	}
	zFile := parent.Output(fileLogger)

	// 创建一个同时输出到控制台和文件的MultiWriter
	mw := io.MultiWriter(zConsole, zFile)
	myLogger := parent.Output(mw)
	return &myLogger, nil
}
