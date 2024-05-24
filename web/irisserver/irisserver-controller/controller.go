package irisserver_controller

import "github.com/kataras/iris/v12"

// 控制器接口
type ApiController interface {
	// 注册路由
	RegisterRoute(irisApp *iris.Application)

	// 路由分组
	apiParty() string
}
