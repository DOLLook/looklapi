package mqconsumers

import "micro-webapi/common/mqutils"

// 初始化消费者
func Initialize() {
	mqutils.BindConsumer()
}
