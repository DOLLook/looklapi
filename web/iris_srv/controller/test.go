package controller

import (
	iris "github.com/kataras/iris/v12"
	"micro-webapi/model/modelbase"
)

type testController struct {
	app *iris.Application
}

var testApi *testController

func init() {
	testApi = &testController{}
	ApiSlice = append(ApiSlice, testApi)
}

func (ctr *testController) apiParty() string {
	return "/test"
}

// 注册路由
func (ctr *testController) RegistRoute(irisApp *iris.Application) {
	ctr.app = irisApp
	party := ctr.app.Party(ctr.apiParty())

	// 绑定路由
	party.Get("/hello", ctr.hello)
	//party.Post("/hello", hello)
}

func (ctr *testController) hello(ctx iris.Context) {
	result := modelbase.NewResponse("hello")
	ctx.JSON(result)
}
