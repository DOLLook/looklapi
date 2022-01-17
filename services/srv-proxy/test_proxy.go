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
	srv srv_isrv.TestSrvInterface `wired:"Autowired"`
}

func init() {
	proxyIns := &testSrvProxy{}
	// 绑定接口映射
	wireutils.Bind(reflect.TypeOf((*srv_isrv.TestSrvInterface)(nil)).Elem(), proxyIns, true, 1)
}

// 代理实现
func (proxy *testSrvProxy) TestLog(log string) error {
	loggers.GetLogger().Debug("before log")

	if err := proxy.srv.TestLog(log); err != nil {
		return err
	}

	loggers.GetLogger().Debug("after log")

	return nil
}
