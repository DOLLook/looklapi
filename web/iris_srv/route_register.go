package iris_srv

import (
	"github.com/kataras/iris/v12"
	"micro-webapi/web/iris_srv/controller"
)

// 注册路由
func registRoute(irisApp *iris.Application) {
	for _, api := range controller.ApiSlice {
		api.RegistRoute(irisApp)
	}
}
