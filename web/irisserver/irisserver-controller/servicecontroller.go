package irisserver_controller

import (
	"github.com/kataras/iris/v12"
	"looklapi/common/appcontext"
	serviceDiscovery "looklapi/common/service-discovery"
	"looklapi/common/wireutils"
	"net/http"
	"reflect"
)

type serviceController struct {
	app              *iris.Application
	appInitCompleted bool
}

func init() {
	serviceApi := &serviceController{}
	wireutils.Bind(reflect.TypeOf((*ApiController)(nil)).Elem(), serviceApi, false, 1)
}

func (ctr *serviceController) apiParty() string {
	return "/service"
}

// 注册路由
func (ctr *serviceController) RegisterRoute(irisApp *iris.Application) {
	ctr.app = irisApp
	party := ctr.app.Party(ctr.apiParty())

	// 绑定路由
	party.Get("/healthCheck", ctr.healthCheck)
}

// 服务健康检查
func (ctr *serviceController) healthCheck(ctx iris.Context) {

	if !ctr.appInitCompleted {
		// 发布程序启动完成消息
		ctr.appInitCompleted = true
		appcontext.GetAppEventPublisher().PublishEvent(appcontext.AppEventInitCompleted(0))
	}

	if serviceDiscovery.GetServiceManager().IsHostCutoff() {
		ctx.StatusCode(http.StatusForbidden)
	} else {
		ctx.StatusCode(http.StatusOK)
	}
	ctx.Next()
}
