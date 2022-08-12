package irisserver_middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/memstore"
	"looklapi/common/utils"
)

// 从context中获取可能存在的controller响应
func getControllerResp(context iris.Context) (ctxStore *memstore.Store, key string, resp interface{}) {
	if context == nil {
		return nil, "", nil
	}

	ctxStore = context.Values()
	if ctxStore == nil {
		return nil, "", nil
	}

	key = utils.ControllerRespContent
	resp = ctxStore.Get(key)
	return
}
