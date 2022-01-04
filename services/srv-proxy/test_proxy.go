package srv_proxy

import (
	"micro-webapi/common/loggers"
	"micro-webapi/common/wireutils"
	_ "micro-webapi/services/srv-impl" // 导入以执行init
	"micro-webapi/services/srv-isrv"
	"reflect"
)

// testsrv 代理
type testSrvProxy struct {
	srv srv_isrv.TestSrvInterface
}

func init() {
	proxyIns := &testSrvProxy{}

	// 注入依赖
	var srvTypeDef *srv_isrv.TestSrvInterface
	proxyIns.srv = wireutils.Resovle(reflect.TypeOf(srvTypeDef)).(srv_isrv.TestSrvInterface)

	// 绑定依赖
	wireutils.Bind(reflect.TypeOf(srvTypeDef), proxyIns, true, 1)
}

// 代理实现
func (proxy *testSrvProxy) TestLog(log string) {
	loggers.GetLogger().Debug("before log")

	proxy.srv.TestLog(log)

	loggers.GetLogger().Debug("after log")
}
