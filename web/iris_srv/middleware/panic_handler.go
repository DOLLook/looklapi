package middleware

import (
	"github.com/kataras/iris/v12"
	"micro-webapi/common/loggers"
	"micro-webapi/model/modelbase"
	"net/http"
)

// 统一panic处理
func PanicHandler() iris.Handler {
	return func(context iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				resp := modelbase.NewErrResponse(err)
				if throws, ok := err.(error); ok {
					loggers.GetLogger().Error(throws)
				}

				context.StopWithJSON(http.StatusOK, resp)
			}
		}()

		context.Next()
	}
}
