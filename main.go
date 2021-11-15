package main

import (
	serviceDiscovery "micro-webapi/common/service-discovery"
	"micro-webapi/mqconsumers"
	"micro-webapi/web"
	"micro-webapi/web/iris_srv"
)

func main() {
	mqconsumers.Initialize()
	serviceDiscovery.GetServiceManager().Init()
	web.LoadLogConfig()
	iris_srv.Start()
}
