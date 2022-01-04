package irisserver_controller

import "github.com/kataras/iris/v12"

var ApiSlice []ApiController

// 控制器接口
type ApiController interface {
	// 注册路由
	RegistRoute(irisApp *iris.Application)

	// 路由分组
	apiParty() string
}
