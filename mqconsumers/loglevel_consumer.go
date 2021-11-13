package mqconsumers

import (
	"go-webapi-fw/common/appcontext"
	"go-webapi-fw/common/loggers"
	"go-webapi-fw/common/mqutils"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/model/modelimpl"
)

/**
日志等级监控器
*/
var logLevelConsumer = mqutils.NewBroadcastConsumer(LOG_LEVEL_CHANGE, 5)

func init() {
	// 消费消息
	logLevelConsumer.Consume = func(msg string) bool {
		var content int32
		if err := utils.JsonToStruct(msg, &content); err != nil {
			loggers.GetLogger().Error(err)
			return false
		}

		logConfig := &modelimpl.ConfigLog{LogLevel: int8(content)}
		appcontext.GetAppEventPublisher().PublishEvent(logConfig)
		return true
	}
}
