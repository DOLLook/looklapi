package main

import (
	serviceDiscovery "micro-webapi/common/service-discovery"
	"micro-webapi/common/wireutils"
	"micro-webapi/mqconsumers"
	_ "micro-webapi/rpc"                // 导入以执行init
	_ "micro-webapi/services/srv-proxy" // 导入以执行init
	"micro-webapi/web"
	"micro-webapi/web/irisserver"
)

func main() {
	wireutils.Inject()
	mqconsumers.Initialize()
	serviceDiscovery.GetServiceManager().Init()
	web.LoadLogConfig()
	irisserver.Start()
}
