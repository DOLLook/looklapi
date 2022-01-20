package irisserver_controller

import (
	"github.com/kataras/iris/v12"
	"micro-webapi/common/appcontext"
	serviceDiscovery "micro-webapi/common/service-discovery"
	"net/http"
)

type serviceController struct {
	app              *iris.Application
	appInitCompleted bool
}

var serviceApi *serviceController

func init() {
	serviceApi = &serviceController{}
	ApiSlice = append(ApiSlice, serviceApi)
}

func (ctr *serviceController) apiParty() string {
	return "/service"
}

// 注册路由
func (ctr *serviceController) RegistRoute(irisApp *iris.Application) {
	ctr.app = irisApp
	party := ctr.app.Party(ctr.apiParty())

	// 绑定路由
	party.Get("/healthCheck", ctr.healthCheck)
}

/**
服务健康检查
*/
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
