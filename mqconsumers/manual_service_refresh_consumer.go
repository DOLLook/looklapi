package mqconsumers

import (
	"go-webapi-fw/common/loggers"
	"go-webapi-fw/common/mqutils"
	serviceDiscovery "go-webapi-fw/common/service-discovery"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/model/modelimpl"
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
