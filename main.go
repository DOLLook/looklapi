package main

import (
	serviceDiscovery "micro-webapi/common/service-discovery"
	"micro-webapi/mqconsumers"
	_ "micro-webapi/rpc" // 导入以执行init
	"micro-webapi/web"
	"micro-webapi/web/irisserver"
)

func main() {
	mqconsumers.Initialize()
	serviceDiscovery.GetServiceManager().Init()
	web.LoadLogConfig()
	irisserver.Start()
}
