package mqconsumers

import (
	"micro-webapi/common/loggers"
	"micro-webapi/common/mqutils"
	serviceDiscovery "micro-webapi/common/service-discovery"
	"micro-webapi/common/utils"
	"micro-webapi/model/modelimpl"
)

/**
服务刷新监控
*/
var manualServiceRefreshConsumer = mqutils.NewBroadcastConsumer(MANUAL_SERVICE_REFRESH, 5)

func init() {
	// 消费消息
	manualServiceRefreshConsumer.Consume = func(msg string) bool {
		var content *modelimpl.ManualService
		if err := utils.JsonToStruct(msg, &content); err != nil {
			loggers.GetLogger().Error(err)
			return false
		}

		serviceDiscovery.GetServiceManager().UpdateManualService(content)
		return true
	}
}
