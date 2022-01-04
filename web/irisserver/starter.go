package irisserver

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/requestid"
	"micro-webapi/common/utils"
	"micro-webapi/config"
	_ "micro-webapi/services/srv-proxy" // 导入以执行init
	"micro-webapi/web/irisserver/irisserver-middleware"
	"sync"
	"time"
)

// 启动web服务
func Start() {
	app := iris.New()
	app.UseRouter(requestid.New())
	app.UseRouter(irisserver_middleware.PanicHandler())
	app.UseRouter(irisserver_middleware.CorsHandler())
	app.DoneGlobal(irisserver_middleware.ErrHandler())
	app.UseError(irisserver_middleware.ErrHandler())

	wg := new(sync.WaitGroup)
	defer wg.Wait()
	iris.RegisterOnInterrupt(func() {
		wg.Add(1)
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		app.Shutdown(ctx)
	})

	registRoute(app)

	cfg := iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		TimeFormat:                        "2006-01-02 15:04:05",
		Charset:                           "UTF-8",
		IgnoreServerErrors:                []string{iris.ErrServerClosed.Error()},
		RemoteAddrHeaders:                 []string{"X-Real-Ip", "X-Forwarded-For"},
	})

	host := utils.HostIp() + ":" + config.AppConfig.Server.Port
	app.Run(iris.Addr(host), cfg)
}
