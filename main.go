package main

import (
	"fmt"
	"go-webapi-fw/common/mongoutils"
	"go-webapi-fw/common/mqutils"
	"go-webapi-fw/common/redisutils"
	"go-webapi-fw/model/modelimpl"
	"go-webapi-fw/web/iris_srv"
)

func main() {
	var configLog = &modelimpl.ConfigLog{LogLevel: 1}
	redisutils.Get("config_log", configLog)

	redisutils.LockAction(func() {
		fmt.Println("加锁测试")
	}, "testlocker", 10)

	mongoutils.Error0("测试一下")

	mqutils.BindConsumer()

	iris_srv.Start()
}
