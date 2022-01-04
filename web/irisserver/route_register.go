package irisserver

import (
	"github.com/kataras/iris/v12"
	"micro-webapi/web/irisserver/irisserver-controller"
)

// 注册路由
func registRoute(irisApp *iris.Application) {
	for _, api := range irisserver_controller.ApiSlice {
		api.RegistRoute(irisApp)
	}
}
