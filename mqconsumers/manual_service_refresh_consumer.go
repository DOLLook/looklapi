package mqconsumers

import (
	"looklapi/common/mqutils"
	serviceDiscovery "looklapi/common/service-discovery"
	"looklapi/model/modelimpl"
	"reflect"
)

// 服务刷新监控器
type manualServiceRefreshConsumer struct {
	messageType reflect.Type
}

func init() {
	consumer := &manualServiceRefreshConsumer{}
	consumer.messageType = reflect.TypeOf((*modelimpl.ManualService)(nil))
	mqutils.NewBroadcastConsumer(mqutils.MANUAL_SERVICE_REFRESH, 5, consumer.messageType, consumer.consume)
}

func (consumer *manualServiceRefreshConsumer) consume(msg interface{}) bool {
	content := msg.(*modelimpl.ManualService)
	serviceDiscovery.GetServiceManager().UpdateManualService(content)
	return true
}
