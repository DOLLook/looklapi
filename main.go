package main

import (
	"looklapi/common/appcontext"
	_ "looklapi/common/service-discovery" // 导入以执行init，无需服务发现可移除
	"looklapi/common/wireutils"
	_ "looklapi/mqconsumers"        // 导入以执行init，无需mq可移除
	_ "looklapi/rpc"                // 导入以执行init，无需rpc可移除
	_ "looklapi/services/srv-proxy" // 导入以执行init
	"looklapi/web/irisserver"
)

func main() {
	wireutils.Inject()
	appcontext.GetAppEventPublisher().PublishEvent(appcontext.AppEventBeanInjected(0))
	irisserver.Start()
}
