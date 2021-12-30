package impl

import (
	"micro-webapi/common/loggers"
	"micro-webapi/common/wireutils"
	"micro-webapi/services/isrv"
	"reflect"
)

type testSrvImpl struct {
}

func init() {
	var srvTypeDef *isrv.TestSrvInterface
	wireutils.Bind(reflect.TypeOf(srvTypeDef), &testSrvImpl{}, false, 1)
}

func (srv *testSrvImpl) TestLog() {
	loggers.GetLogger().Debug("test log")
}
