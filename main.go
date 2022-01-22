package main

import (
	"micro-webapi/common/appcontext"
	_ "micro-webapi/common/service-discovery" // 导入以执行init，无需服务发现可移除
	"micro-webapi/common/wireutils"
	_ "micro-webapi/mqconsumers"        // 导入以执行init，无需mq可移除
	_ "micro-webapi/rpc"                // 导入以执行init，无需rpc可移除
	_ "micro-webapi/services/srv-proxy" // 导入以执行init
	"micro-webapi/web/irisserver"
)

func main() {
	wireutils.Inject()
	appcontext.GetAppEventPublisher().PublishEvent(appcontext.AppEventBeanInjected(0))
	irisserver.Start()
}
