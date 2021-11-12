package mqconsumers

import "go-webapi-fw/common/mqutils"

// 初始化消费者
func Initialize() {
	mqutils.BindConsumer()
}
