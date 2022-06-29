package mqconsumers

import (
	"looklapi/common/appcontext"
	"looklapi/common/loggers"
	"looklapi/common/mqutils"
	"reflect"
)

// 日志等级监控器
type logLevelConsumer struct {
	messageType reflect.Type
}

func init() {
	consumer := &logLevelConsumer{}
	consumer.messageType = reflect.TypeOf((*int)(nil)).Elem()
	mqutils.NewBroadcastConsumer(mqutils.LOG_LEVEL_CHANGE, 5, consumer.messageType, consumer.consume)
}

func (consumer *logLevelConsumer) consume(msg interface{}) bool {
	content := msg.(int)
	logConfig := &loggers.ConfigLog{LogLevel: int8(content)}
	appcontext.GetAppEventPublisher().PublishEvent(logConfig)
	return true
}
