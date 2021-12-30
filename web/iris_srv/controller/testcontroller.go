package controller

import (
	iris "github.com/kataras/iris/v12"
	"micro-webapi/common/wireutils"
	"micro-webapi/model/modelbase"
	"micro-webapi/services/isrv"
	"reflect"
)

type testController struct {
	app     *iris.Application
	testSrv isrv.TestSrvInterface
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

	// 注入依赖
	var srvTypeDef *isrv.TestSrvInterface
	ctr.testSrv = wireutils.Resovle(reflect.TypeOf(srvTypeDef)).(isrv.TestSrvInterface)

	// 绑定路由
	party.Get("/hello", ctr.hello)
	//party.Post("/hello", hello)
}

func (ctr *testController) hello(ctx iris.Context) {
	ctr.testSrv.TestLog() // 执行代理
	result := modelbase.NewResponse("hello")
	ctx.JSON(result)
	ctx.Next()
}
