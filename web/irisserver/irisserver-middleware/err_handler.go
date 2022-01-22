package irisserver_middleware

import (
	"github.com/kataras/iris/v12"
	"looklapi/common/loggers"
	"looklapi/errs"
	"looklapi/model/modelbase"
	"net/http"
)

// 统一异常处理
func ErrHandler() iris.Handler {
	return func(context iris.Context) {

		if err := context.GetErr(); err != nil {
			var bErr error
			bErr, ok := err.(*errs.BllError)
			if !ok {
				bErr = errs.NewBllError(err.Error())
			}

			loggers.GetLogger().Error(bErr)
			resp := modelbase.NewErrResponse(bErr)
			context.StopWithJSON(http.StatusOK, resp)
		}

		context.Next()
	}
}
