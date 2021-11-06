package middleware

import (
	"github.com/kataras/iris/v12"
	"go-webapi-fw/common/mongoutils"
)

// 统一异常处理
func ExceptionHandler() iris.Handler {
	return func(context iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				if throws, ok := err.(error); ok {
					mongoutils.Error(throws)
				}

				if context.IsStopped() {
					return
				}
			}
		}()

		context.Next()
	}
}
