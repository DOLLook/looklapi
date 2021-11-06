package iris_srv

import (
	"github.com/kataras/iris/v12"
	"go-webapi-fw/web/iris_srv/controller"
)

// 注册路由
func registRoute(irisApp *iris.Application) {
	for _, api := range controller.ApiSlice {
		api.RegistRoute(irisApp)
	}
}
