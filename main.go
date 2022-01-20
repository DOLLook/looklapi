package main

import (
	"micro-webapi/common/appcontext"
	serviceDiscovery "micro-webapi/common/service-discovery"
	"micro-webapi/common/wireutils"
	_ "micro-webapi/mqconsumers"        // 导入以执行init
	_ "micro-webapi/rpc"                // 导入以执行init
	_ "micro-webapi/services/srv-proxy" // 导入以执行init
	"micro-webapi/web"
	"micro-webapi/web/irisserver"
)

func main() {
	wireutils.Inject()
	appcontext.GetAppEventPublisher().PublishEvent(appcontext.AppEventBeanInjected(0))
	serviceDiscovery.GetServiceManager().Init()
	web.LoadLogConfig()
	irisserver.Start()
}
