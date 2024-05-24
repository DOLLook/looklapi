package irisserver

import (
	"github.com/kataras/iris/v12"
	"looklapi/common/wireutils"
	irisserver_controller "looklapi/web/irisserver/irisserver-controller"
	"reflect"
)

// 注册路由
func registerRoute(irisApp *iris.Application) {
	for _, ctr := range wireutils.ResovleAll(reflect.TypeOf((*irisserver_controller.ApiController)(nil)).Elem()) {
		if api, ok := ctr.(irisserver_controller.ApiController); ok {
			api.RegisterRoute(irisApp)
		}
	}
}
