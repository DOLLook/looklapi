package main

import (
	"go-webapi-fw/common/mqutils"
	serviceDiscovery "go-webapi-fw/common/service-discovery"
	"go-webapi-fw/mqconsumers"
	"go-webapi-fw/web/iris_srv"
)

func main() {
	mqconsumers.Initialize()
	mqutils.BindConsumer()
	serviceDiscovery.Register()
	serviceDiscovery.GetServiceManager().Init()
	iris_srv.Start()
}
