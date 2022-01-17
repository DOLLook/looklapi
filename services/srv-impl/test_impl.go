package srv_impl

import (
	"micro-webapi/common/loggers"
	"micro-webapi/common/wireutils"
	"micro-webapi/services/srv-isrv"
	"reflect"
)

type testSrvImpl struct {
}

func init() {
	testSrv := &testSrvImpl{}
	// 绑定接口映射
	wireutils.Bind(reflect.TypeOf((*srv_isrv.TestSrvInterface)(nil)).Elem(), testSrv, false, 1)
}

func (srv *testSrvImpl) TestLog(log string) error {
	loggers.GetLogger().Debug("test log: " + log)
	return nil
}
