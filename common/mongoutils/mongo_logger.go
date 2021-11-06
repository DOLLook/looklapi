package mongoutils

import (
	"go-webapi-fw/common/utils"
	"go-webapi-fw/config"
	"go-webapi-fw/errs"
	"runtime/debug"
	"strconv"
	"strings"
)

var _LOGLEVEL = _ERROR

const (
	_OFF = iota
	_FATAL
	_ERROR
	_WARN
	_INFO
	_DEBUG
	_ALL
)

// 更新日志等级
func RefreshLogLevel(loglevel int32) {
	_LOGLEVEL = int(loglevel)
}

// 调试日志
func Debug(msg string) {
	if _LOGLEVEL < _DEBUG {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := createLog(msg, _DEBUG)
	go insert(log)
}

// 提示
func Info(msg string) {
	if _LOGLEVEL < _INFO {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := createLog(msg, _INFO)
	go insert(log)
}

// 警告
func Warn(msg string) {
	if _LOGLEVEL < _WARN {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := createLog(msg, _WARN)
	go insert(log)
}

// 错误日志
func Error0(msg string) {
	if _LOGLEVEL < _ERROR {
		return
	}

	if utils.IsEmpty(msg) {
		return
	}

	log := createLog(msg, _ERROR)
	go insert(log)
}

// 错误日志
func Error(err error) {
	if _LOGLEVEL < _ERROR {
		return
	}

	if err == nil {
		return
	}

	var msg string
	if bErr, ok := err.(*errs.BllError); ok {
		msg = bErr.Msg
	} else {
		msg = err.Error()
	}

	log := createLog(msg, _ERROR)
	go insert(log)
}

func createLog(msg string, level int32) *systemRuntineLog {
	log := NewMongoLog()
	log.Instance = config.AppConfig.Server.Name
	log.HostIp = utils.HostIp()
	log.Content = msg
	log.Level = level
	stackStr := string(debug.Stack())

	stackSlice := strings.Split(stackStr, "\n")
	if level == _ERROR || level == _FATAL {
		var temp []string
		temp = append(temp, stackSlice[0])
		temp = append(temp, stackSlice[7:]...)
		log.Stacktrace = strings.Join(temp, "\n")
	}

	if routineId, err := strconv.Atoi(strings.Split(stackSlice[0], " ")[1]); err == nil {
		log.ThreadId = int32(routineId)
	}

	//log.ClassName

	return log
}

func insert(log *systemRuntineLog) {
	GetCollection(log.TbCollName()).InsertOne(nil, log)
}
