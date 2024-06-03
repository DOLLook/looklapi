package srv_isrv

// 测试接口
type TestSrvInterface interface {
	TestLog(log string) error
	TestLogProxyVersion(log string) error
}
