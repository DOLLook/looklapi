# micro-webapi
micro-webapi是一个为业务构建的基于kataras/iris的微服务开发脚手架项目。
* 项目集成了mongodb，redis，并实现了redis基本操作和分布式锁操作。
* 集成rabbitmq作为消息中间件，实现了rabbitmq连接池，以及消息重试策略，可以根据业务快速定义消费者。
* 集成consul服务发现，实现了服务注册、健康检查和健康服务缓存刷新，通过简单配置即可进行服务注册和获取。
* 内置基于事件的应用消息发布工具AppEventPublisher和AppObserver。
* 内置轻量化ioc容器实现依赖注入及aop切面编程。
* controller api参数的自动映射，以及可选的参数校验器，快速构建web业务。
* 基于fasthttp的声明式http调用客户端，只需简单声明即可完成远程调用，配合服务发现，快速构建服务间调用。

项目不限制web框架，开发者也可以方便地接入gin等web框架处理请求。

## 1. controller
### 1.1 api路由绑定
* controller api接口
* paramValidator 参数校验器（可选）
* beforeHandlers 接口请求前拦截器（可选）
* afterHandlers 接口请求后拦截器（可选）

        func RegisterController(irisApp *iris.Application, apiParty string, routePath string, httpMethod string, controller interface{}, paramValidator interface{}, beforeHandlers []iris.Handler, afterHandlers []iris.Handler)
        
* 注册接口

        irisserver_middleware.RegisterController(
                ctr.app,
                ctr.apiParty(),
                "/testLog",
                http.MethodGet,
                ctr.testLog,
                ctr.testLogParamValidator,
                nil,
                nil)

### 1.2 api参数自动映射
* 无需解析请求参数，自动映射到接口
```
package irisserver_controller
import "fmt"
import "errors"

type testController struct {
	
}

func (ctr *testController) testLog(log string) error {
    fmt.Println(log)
    return nil
}

type LogBody struct{
    id      int
    content string
}

// 结构化参数
func (ctr *testController) testLog(log *LogBody) error {
    fmt.Println(log)
    return nil
}

// 可选的参数校验器

// testLog参数校验
func (ctr *testController) testLogParamValidator(log string) error {
    if len(log)==0{
        return errors.New("invalid param")
    }

	return nil
}

```

### 2. 依赖注入
采用无第三方依赖的内置ioc容器实现
* 容器实现了类型与实例绑定，代理注入(暂未支持多层代理)
* itype 待绑定的类型
* target 绑定的实例
* proxy 指定是否为代理
* priority 指定注入优先级, 在自动注入时按优先级注入
```
func Bind(itype reflect.Type, target interface{}, proxy bool, priority int)
```

### 3. service层注入
* 定义接口
```
package srv_isrv

type TestSrvInterface interface {
	TestLog(log string) error
}
```
* 定义实现
```
package srv_impl

// 定义实现
type testSrvImpl struct {
}

func init() {
	testSrv := &testSrvImpl{}
	// 绑定接口映射
	wireutils.Bind(reflect.TypeOf((*srv_isrv.TestSrvInterface)(nil)).Elem(), testSrv, false, 1)
}

// 实现接口
func (srv *testSrvImpl) TestLog(log string) error {
	fmt.Println(log)
	return nil
}
```
* controller注入。使用Tag `wired:"Autowired"`自动注入
```
package irisserver_controller

import (
    "srv-isrv"
    "reflect"
)

type testController struct {
	testSrv srv_isrv.TestSrvInterface `wired:"Autowired"`
}

var testApi *testController

func init() {
	testApi = &testController{}
	wireutils.Bind(reflect.TypeOf((*testController)(nil)).Elem(), testApi, false, 1)
}

func (ctr *testController) testLog(log string) error {
    // 使用注入的testSrv
	return ctr.testSrv.TestLog(log)
}

```

#### service层aop  
在需要时，可通过可选的代理注入实现aop
* 定义代理并实现与业务层相同接口，代理层在容器中具有最高优先级，在注入时会优先注入
```
package srv_proxy

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
	fmt.Println("before log")

	if err := proxy.srv.TestLog(log); err != nil {
		return err
	}

	fmt.Println("after log")

	return nil
}
```

### 4. 声明式http调用
* 声明调用接口
```
type RpcService interface {
	SrvName() string
}

// TestService为远程测试服务，暴露接口TestApi
type TestService struct {

	// 测试接口
	// header 请求头，可选
	// body 自定义参数, 可选
	// temp1, temp2 自定义url参数, 可选
	// resultPtr 自定义的请求结果接收指针，必传
	// tag route: 指定请求路由, method: 指定求请方式, alias: 指定url参数别名
	TestApi func(header http.Header, body []int, temp1 string, temp2 int, resultPtr *modelbase.ResponseResult) error `route:"/api/testapi" method:"POST" alias:"[temp1,temp2]"`
}

// 固定实现
func (srv *TestService) SrvName() string {
	return "YOUR_SERVICE_NAME"
}

// 初始化注册接口
func init() {
	register(&TestService{})
}
```

* 调用接口
```
// api错误响应
type errResponse struct {
	// 是否成功
	IsSuccess bool
	// 错误码
	ErrorCode int
	// 错误信息
	ErrorMsg string
}

// api请求响应值
type ResponseResult struct {
	errResponse
	// 结果
	Result interface{}
}

// 请求结果
func NewResponse(data interface{}) (result *ResponseResult) {
	result = &ResponseResult{
		Result: data,
	}
	result.IsSuccess = true
	return
}

// 调用
func testRequest(){
    resp := NewResponse(make(map[int]int))
    testClient := rpc.GetHttpRpcClient("YOUR_SERVICE_NAME").(*TestService)
    if err := testClient.TestApi(nil,[]int{1,2,3},"temp1",0, resp); err != nil {
        
    }
}

```

### 5. MQ
基于rabbitmq实现了两类消费者，worker模式和broadcast模式
* worker模式  
maxRetry 消息最大重试次数, 重试过程中须自行保证消息幂等性
```
func NewWorkQueueConsumer(routeKey string, concurrency uint32, prefetchCount uint32, parallel bool, maxRetry uint32) *consumer
```
* broadcast模式  
maxRetry 消息最大重试次数, 重试过程中须自行保证消息幂等性
```
func NewBroadcastConsumer(exchange string, maxRetry uint32) *consumer
```

* 示例
```
// 创建消费者
var testConsumer = mqutils.NewBroadcastConsumer("your_exchange", 5)

// 初始化
func init() {
	// 消费消息
	testConsumer.Consume = func(msg string) bool {
		return true
	}
}
```