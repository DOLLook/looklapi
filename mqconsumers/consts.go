package mqconsumers

// 交换器枚举
type RabbitMqExchange = string

// 路由枚举
type RabbitMqRouteKey = string

const (
	LOG_LEVEL_CHANGE       RabbitMqExchange = "log_level_change"       // 日志等级变更交换器
	MANUAL_SERVICE_REFRESH RabbitMqExchange = "manual_service_refresh" // 服务配置刷新交换器
)
