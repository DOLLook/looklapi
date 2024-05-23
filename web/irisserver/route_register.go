package irisserver

import (
	"github.com/kataras/iris/v12"
	"looklapi/web/irisserver/irisserver-controller"
)

// 注册路由
func registerRoute(irisApp *iris.Application) {
	for _, api := range irisserver_controller.ApiSlice {
		api.RegisterRoute(irisApp)
	}
}
