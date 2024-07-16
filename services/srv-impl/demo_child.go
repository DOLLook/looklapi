package srv_impl

import (
	"looklapi/common/loggers"
	"looklapi/common/wireutils"
	srv_isrv "looklapi/services/srv-isrv"
	"reflect"
)

type demoChild struct {
	demoParent `wired:"Autowired"`
}

func init() {
	srv := &demoChild{}
	srv.virtualFunc = srv.virtual

	// 绑定接口映射
	wireutils.Bind(reflect.TypeOf((*srv_isrv.InheritTestInterface)(nil)).Elem(), srv, false, 1)
}

func (srv *demoChild) TestInherit(log string) {
	srv.commonMethod(log)
}

func (srv *demoChild) virtual(log string) {
	loggers.GetLogger().Info("child do")
	srv.testSrv.TestLog(log)
}
