package iris_srv

import (
	"context"
	"github.com/kataras/iris/v12"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/config"
	"go-webapi-fw/web/iris_srv/middleware"
	"sync"
	"time"
)

// 启动web框架
func Start() {
	app := iris.New()
	//app.Use(middleware.ExceptionHandler())
	app.UseError(middleware.ExceptionHandler())
	app.UseRouter(middleware.CorsHandler())

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
