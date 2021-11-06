package controller

import (
	iris "github.com/kataras/iris/v12"
	"go-webapi-fw/model/modelbase"
)

var TestApi *testController

type testController struct {
	app *iris.Application
}

func init() {
	TestApi = &testController{}
	ApiSlice = append(ApiSlice, TestApi)
}

// 注册路由
func (test *testController) RegistRoute(irisApp *iris.Application) {
	TestApi.app = irisApp
	// 绑定路由
	party := TestApi.app.Party("/test")
	//party.Post("/hello", hello)
	party.Get("/hello", hello)
}

func hello(ctx iris.Context) {
	result := modelbase.NewResponse("hello")
	ctx.JSON(result)
}
