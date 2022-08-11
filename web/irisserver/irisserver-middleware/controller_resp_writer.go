package irisserver_middleware

import (
	"github.com/kataras/iris/v12"
	"looklapi/common/utils"
)

// 写入controller响应
func ControllerRespWriter() iris.Handler {
	return func(context iris.Context) {
		ctxStore := context.Values()
		if ctxStore.Exists(utils.ControllerRespContent) {
			resp := ctxStore.Get(utils.ControllerRespContent)
			if _, err := context.JSON(resp); err != nil {
				context.SetErr(err)
			}
			ctxStore.Remove(utils.ControllerRespContent)
		}

		context.Next()
	}
}
