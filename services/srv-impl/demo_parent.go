package srv_impl

import (
	"looklapi/common/loggers"
	srv_isrv "looklapi/services/srv-isrv"
)

type demoParent struct {
	testSrv     srv_isrv.TestSrvInterface `wired:"Autowired"`
	virtualFunc func(string)
}

func (srv demoParent) commonMethod(log string) {
	loggers.GetLogger().Info("parent start")
	srv.virtualFunc(log)
	loggers.GetLogger().Info("parent end")
}
