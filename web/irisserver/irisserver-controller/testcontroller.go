package irisserver_controller

import (
	iris "github.com/kataras/iris/v12"
	"micro-webapi/common/utils"
	"micro-webapi/common/wireutils"
	"micro-webapi/errs"
	"micro-webapi/services/srv-isrv"
	irisserver_middleware "micro-webapi/web/irisserver/irisserver-middleware"
	"net/http"
	"reflect"
)

type testController struct {
	app     *iris.Application
	testSrv srv_isrv.TestSrvInterface
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

	// 注入依赖
	var srvTypeDef *srv_isrv.TestSrvInterface
	ctr.testSrv = wireutils.Resovle(reflect.TypeOf(srvTypeDef)).(srv_isrv.TestSrvInterface)

	// 绑定路由
	irisserver_middleware.RegisterController(
		ctr.app,
		ctr.apiParty(),
		"/hello",
		http.MethodGet,
		ctr.testSrv.TestLog,
		ctr.testLogParamValidator)
}

// testLog参数校验
func (ctr *testController) testLogParamValidator(log string) error {
	if utils.IsEmpty(log) {
		return errs.NewBllError("参数错误")
	}

	return nil
}
