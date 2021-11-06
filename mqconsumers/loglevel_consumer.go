package mqconsumers

import (
	"go-webapi-fw/common/mongoutils"
	"go-webapi-fw/common/mqutils"
	"go-webapi-fw/common/utils"
)

/**
日志等级监控器
*/
var logLevelConsumer = mqutils.NewBroadcastConsumer(LOG_LEVEL_CHANGE, 5)

func init() {
	logLevelConsumer.Consume = consume
}

// 消费消息
func consume(msg string) bool {
	var content int32
	if err := utils.JsonToStruct(msg, &content); err != nil {
		mongoutils.Error(err)
		return false
	}

	mongoutils.RefreshLogLevel(content)

	return false
}
