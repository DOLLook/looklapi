package irisserver_middleware

import (
	"github.com/kataras/iris/v12"
)

// 写入controller响应
func ControllerRespWriter() iris.Handler {
	return func(context iris.Context) {
		ctxStore, key, resp := getControllerResp(context)
		if resp != nil {
			if _, err := context.JSON(resp); err != nil {
				context.SetErr(err)
			}
			ctxStore.Remove(key)
		}

		context.Next()
	}
}
