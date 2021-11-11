package controller

import (
	"github.com/kataras/iris/v12"
	serviceDiscovery "go-webapi-fw/common/service-discovery"
	"net/http"
)

type serviceController struct {
	app *iris.Application
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
	if serviceDiscovery.GetServiceManager().IsHostCutoff() {
		ctx.StatusCode(http.StatusForbidden)
	} else {
		ctx.StatusCode(http.StatusOK)
	}
}
