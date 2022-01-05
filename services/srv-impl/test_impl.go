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
	var srvTypeDef *srv_isrv.TestSrvInterface
	wireutils.Bind(reflect.TypeOf(srvTypeDef), &testSrvImpl{}, false, 1)
}

func (srv *testSrvImpl) TestLog(log string) error {
	loggers.GetLogger().Debug("test log: " + log)
	return nil
}
