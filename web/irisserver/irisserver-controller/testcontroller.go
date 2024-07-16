package irisserver_controller

import (
	"context"
	"looklapi/common/utils"
	"looklapi/common/wireutils"
	"looklapi/errs"
	"looklapi/model/modelbase"
	srv_isrv "looklapi/services/srv-isrv"
	irisserver_middleware "looklapi/web/irisserver/irisserver-middleware"
	"net/http"
	"reflect"

	"github.com/kataras/iris/v12"
)

type testController struct {
	app            *iris.Application
	testSrv        srv_isrv.TestSrvInterface     `wired:"Autowired"`
	inheritTestSrv srv_isrv.InheritTestInterface `wired:"Autowired"`
}

func init() {
	testApi := &testController{}
	wireutils.Bind(reflect.TypeOf((*ApiController)(nil)).Elem(), testApi, false, 1)
}

func (ctr *testController) apiParty() string {
	return "/test"
}

// 注册路由
func (ctr *testController) RegisterRoute(irisApp *iris.Application) {
	ctr.app = irisApp

	//// 注入依赖
	//ctr.testSrv = wireutils.Resovle(reflect.TypeOf((*srv_isrv.TestSrvInterface)(nil)).Elem()).(srv_isrv.TestSrvInterface)

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

	// 绑定路由
	irisserver_middleware.RegisterController(
		ctr.app,
		ctr.apiParty(),
		"/proxy",
		http.MethodGet,
		ctr.testLogProxy,
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

	irisserver_middleware.RegisterController(
		ctr.app,
		ctr.apiParty(),
		"/inherit",
		http.MethodGet,
		ctr.inheritTest,
		nil,
		nil,
		nil)
}

// test log api
func (ctr *testController) testLog(log string) error {
	return ctr.testSrv.TestLog(log)
}

// test log api
func (ctr *testController) testLogProxy(log string) error {
	return ctr.testSrv.TestLogProxyVersion(log)
}

// test log with result and context api
func (ctr *testController) testLogWithResult(ctx context.Context, log string) (*modelbase.ResponseResult, error) {
	if err := ctr.testSrv.TestLog(log); err != nil {
		return nil, err
	}

	httpHeader := utils.GetHttpHeader(ctx)
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
func (ctr *testController) testLogWithResultParamValidator(ctx context.Context, log string) error {
	if utils.IsEmpty(log) {
		return errs.NewBllError("参数错误")
	}

	return nil
}

// inherit test api
func (ctr *testController) inheritTest(log string) error {
	ctr.inheritTestSrv.TestInherit(log)
	return nil
}
