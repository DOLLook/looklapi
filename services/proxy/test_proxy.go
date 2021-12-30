package proxy

import (
	"micro-webapi/common/loggers"
	"micro-webapi/common/wireutils"
	_ "micro-webapi/services/impl" // 导入以执行init
	"micro-webapi/services/isrv"
	"reflect"
)

// testsrv 代理
type testSrvProxy struct {
	srv isrv.TestSrvInterface
}

func init() {
	proxyIns := &testSrvProxy{}

	// 注入依赖
	var srvTypeDef *isrv.TestSrvInterface
	proxyIns.srv = wireutils.Resovle(reflect.TypeOf(srvTypeDef)).(isrv.TestSrvInterface)

	// 绑定依赖
	wireutils.Bind(reflect.TypeOf(srvTypeDef), proxyIns, true, 1)
}

// 代理实现
func (proxy *testSrvProxy) TestLog() {
	loggers.GetLogger().Debug("before log")

	proxy.srv.TestLog()

	loggers.GetLogger().Debug("after log")
}
