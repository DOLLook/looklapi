package main

import (
	serviceDiscovery "go-webapi-fw/common/service-discovery"
	"go-webapi-fw/mqconsumers"
	"go-webapi-fw/web/iris_srv"
)

func main() {
	mqconsumers.Initialize()
	serviceDiscovery.GetServiceManager().Init()
	iris_srv.Start()
}
