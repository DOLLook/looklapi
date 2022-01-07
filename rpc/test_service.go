package rpc

import (
	"micro-webapi/common/utils"
	"micro-webapi/model/modelbase"
	"net/http"
)

// 测试服务
type TestService struct {

	// 测试接口
	TestApi func(header http.Header, body []int, temp1 string, temp2 int, resultPtr *modelbase.ResponseResult) error `route:"/api/testapi" method:"POST" alias:"[temp1,temp2]"`
}

func (srv *TestService) SrvName() string {
	return string(utils.TEST_SERVICE)
}

func init() {
	register(&TestService{})
}
