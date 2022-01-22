package irisserver_middleware

import (
	"errors"
	"github.com/kataras/iris/v12"
	"looklapi/common/loggers"
	"looklapi/model/modelbase"
	"net/http"
)

// 统一panic处理
func PanicHandler() iris.Handler {
	return func(context iris.Context) {
		defer func() {
			if err := recover(); err != nil {

				if throws, ok := err.(error); ok {
					loggers.GetLogger().Error(throws)
				} else if msg, ok := err.(string); ok {
					loggers.GetLogger().Error(errors.New(msg))
				}

				resp := modelbase.NewErrResponse(err)
				context.StopWithJSON(http.StatusOK, resp)
			}
		}()

		context.Next()
	}
}
