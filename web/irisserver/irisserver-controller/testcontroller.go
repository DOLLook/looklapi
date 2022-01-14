package irisserver_controller

import (
	"context"
	iris "github.com/kataras/iris/v12"
	"micro-webapi/common/utils"
	"micro-webapi/common/wireutils"
	"micro-webapi/errs"
	"micro-webapi/model/modelbase"
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
		ctr.testLog,
		ctr.testLogParamValidator,
		nil,
		nil)

	irisserver_middleware.RegisterController(
		ctr.app,
		ctr.apiParty(),
		"/hello1",
		http.MethodGet,
		ctr.testLogWithResult,
		ctr.testLogWithResultParamValidator,
		nil,
		nil)
}

// test log api
func (ctr *testController) testLog(log string) error {
	return ctr.testSrv.TestLog(log)
}

// test log with result and context api
func (ctr *testController) testLogWithResult(context context.Context, log string) (*modelbase.ResponseResult, error) {
	if err := ctr.testSrv.TestLog(log); err != nil {
		return nil, err
	}

	httpHeader := context.Value(utils.HttpRequestHeader).(http.Header)
	return modelbase.NewResponse(utils.StructToJson(httpHeader)), nil
}

// testLog参数校验
func (ctr *testController) testLogParamValidator(log string) error {
	if utils.IsEmpty(log) {
		return errs.NewBllError("参数错误")
	}

	return nil
}

// testLog参数校验
func (ctr *testController) testLogWithResultParamValidator(context context.Context, log string) error {
	if utils.IsEmpty(log) {
		return errs.NewBllError("参数错误")
	}

	return nil
}
