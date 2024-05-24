package rpc

import (
	"looklapi/model/modelbase"
	"net/http"
)

// 测试服务
type TestService struct {

	// 测试接口
	// header 请求头，可选
	// body 自定义参数, 可选
	// temp1, temp2 自定义url参数, 可选
	// resultPtr 自定义的请求结果接收指针，必传
	// tag route: 指定请求路由, method: 指定求请方式, alias: 指定url参数别名
	TestApi func(header http.Header, body []int, temp1 string, temp2 int, resultPtr *modelbase.ResponseResult) error `route:"/api/testapi" method:"POST" alias:"[temp1,temp2]"`
}

func (srv *TestService) SrvName() string {
	// when you use service discovery, define your service name here. And the request send by framework will find the healthy endpoint of the service.
	return "your-service-name"
}

func (srv *TestService) EndPoint() string {
	// when you not use service discovery, specify a constant endpoint for request.
	// notice: if you use endpoint, the SrvName set for service discovery will be ignored.
	return "https://your.service.com"
}

func (srv *TestService) RequestInterceptor() []interceptor {
	return nil
}

func init() {
	register(&TestService{})
}
