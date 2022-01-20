package mqconsumers

import (
	"micro-webapi/common/appcontext"
	"micro-webapi/common/mqutils"
	"micro-webapi/model/modelimpl"
	"reflect"
)

// 日志等级监控器
type logLevelConsumer struct {
	messageType reflect.Type
}

func init() {
	consumer := &logLevelConsumer{}
	consumer.messageType = reflect.TypeOf((*int)(nil)).Elem()
	mqutils.NewBroadcastConsumer(LOG_LEVEL_CHANGE, 5, consumer.messageType, consumer.consume)
}

func (consumer *logLevelConsumer) consume(msg interface{}) bool {
	content := msg.(int)
	logConfig := &modelimpl.ConfigLog{LogLevel: int8(content)}
	appcontext.GetAppEventPublisher().PublishEvent(logConfig)
	return true
}
