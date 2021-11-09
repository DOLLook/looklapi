package mqutils

import (
	"errors"
	"github.com/streadway/amqp"
	"go-webapi-fw/common/mongoutils"
	"go-webapi-fw/common/utils"
)

/**
绑定消费者
*/
func BindConsumer() {
	if _hasConsumerBind || _consumerContainer == nil || len(_consumerContainer) < 1 {
		return
	}

	for _, item := range _consumerContainer {
		bindConsumer(item)
	}
	_hasConsumerBind = true
}

/**
绑定消费者
*/
func bindConsumer(consumer *consumer) {
	if consumer == nil {
		panic("invalid nil consumer")
	}

	switch consumer.Type {
	case _WorkQueue:
		for i := uint32(0); i < consumer.Concurrency; i++ {
			bindWorkQueueConsumer(consumer)
		}
		break
	case _Broadcast:
		bindBroadcastConsumer(consumer)
		break
	}
}

/**
绑定工作队列消费者

consumer 消费者
*/
func bindWorkQueueConsumer(consumer *consumer) {
	if consumer == nil {
		panic("invalid nil consumer")
	}

	if consumer.Type != _WorkQueue {
		panic("invalid consumer type")
	}

	if consumer.Consume == nil {
		panic("must bind consume func")
	}

	recChan := getConsumerChannel()
	if _, err := recChan.Channel.QueueDeclare(consumer.RouteKey, true, false, false, false, nil); err != nil {
		panic(err)
	}

	if err := recChan.Channel.Qos(int(consumer.PrefetchCount), 0, false); err != nil {
		panic(err)
	}

	deliverCh, err := recChan.Channel.Consume(consumer.RouteKey, "", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		//for delivery := range deliverCh {
		//	if consumer.Parallel && consumer.PrefetchCount > 1 {
		//		go consume(&delivery, consumer)
		//	} else {
		//		consume(&delivery, consumer)
		//	}
		//}
		for {
			select {
			case delivery := <-deliverCh:
				if consumer.Parallel && consumer.PrefetchCount > 1 {
					go consume(&delivery, consumer)
				} else {
					consume(&delivery, consumer)
				}
			case err := <-errChan:
				mongoutils.Error(err)
				removeConsumerChannel(recChan)
				bindConsumer(consumer)
				return
			}
		}
	}()

	addConsumerChannel(recChan)
}

/**
绑定广播队列消费者

consumer 消费者
*/
func bindBroadcastConsumer(consumer *consumer) {
	if consumer == nil {
		panic("invalid nil consumer")
	}

	if consumer.Type != _Broadcast {
		panic("invalid consumer type")
	}

	if consumer.Consume == nil {
		panic("must bind consume func")
	}

	recChan := getConsumerChannel()
	if err := recChan.Channel.ExchangeDeclare(consumer.Exchange, "fanout", false, true, false, false, nil); err != nil {
		panic(err)
	}

	queue, err := recChan.Channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		panic(err)
	}

	if err := recChan.Channel.QueueBind(queue.Name, "", consumer.Exchange, false, nil); err != nil {
		panic(err)
	}

	deliverCh, err := recChan.Channel.Consume(queue.Name, "", false, true, false, false, nil)
	if err != nil {
		panic(err)
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		//for delivery := range deliverCh {
		//	consume(&delivery, consumer)
		//}
		for {
			select {
			case delivery := <-deliverCh:
				consume(&delivery, consumer)
			case err := <-errChan:
				mongoutils.Error(err)
				removeConsumerChannel(recChan)
				bindConsumer(consumer)
				return
			}
		}
	}()

	addConsumerChannel(recChan)
}

func consume(delivery *amqp.Delivery, consumer *consumer) {
	content := string(delivery.Body)
	if consumer.onRecieved(content) {
		delivery.Ack(false)
	} else {
		if err := delivery.Nack(false, true); err != nil {
			mongoutils.Error(errors.New("消息Nack失败: " + content))
		}
	}
}

// 消费者类型
type consumerType = uint8

// 消费者类型
const (
	_Invalid   consumerType = iota
	_WorkQueue              // 工作队列消费者
	_Broadcast              // 广播消费者
)

var _hasConsumerBind bool          // 消费者是否已绑定
var _consumerContainer []*consumer // 消费者容器

// 消费者
type consumer struct {
	Type     consumerType      // 消费者类型
	MaxRetry uint32            // 最大重试次数
	Consume  func(string) bool // 处理器

	Exchange string // broadcast交换器名称

	RouteKey      string // workqueue路由地址
	Concurrency   uint32 // workqueue并发消费者数量
	PrefetchCount uint32 // workqueue从队列中同时deliver的消息数量
	Parallel      bool   // workqueue是否开启并行消费
}

// 新建消费者
func NewWorkQueueConsumer(routeKey string, concurrency uint32, prefetchCount uint32, parallel bool, maxRetry uint32) *consumer {
	if utils.IsEmpty(routeKey) {
		panic("invalid routekey")
	}

	if concurrency < 1 {
		panic("workqueue consumer concurrency must greater than 0")
	}

	if prefetchCount < 1 {
		panic("workqueue consumer prefetchCount must greater than 0")
	}

	if maxRetry < 1 {
		panic("workqueue consumer maxRetry must greater than 0")
	}

	cs := &consumer{
		Type:          _WorkQueue,
		MaxRetry:      maxRetry,
		RouteKey:      routeKey,
		Concurrency:   concurrency,
		PrefetchCount: prefetchCount,
		Parallel:      parallel,
	}

	_consumerContainer = append(_consumerContainer, cs)
	return cs
}

// 新建消费者
func NewBroadcastConsumer(exchange string, maxRetry uint32) *consumer {
	if utils.IsEmpty(exchange) {
		panic("invalid exchange")
	}

	if maxRetry < 1 {
		panic("broadcast consumer maxRetry must greater than 0")
	}

	cs := &consumer{
		Type:     _Broadcast,
		MaxRetry: maxRetry,
		Exchange: exchange,
	}

	_consumerContainer = append(_consumerContainer, cs)
	return cs
}

// 接收到消息
func (consumer *consumer) onRecieved(msg string) bool {
	if consumer.Type == _WorkQueue {
		// todo 服务健康检查
	}

	if utils.IsEmpty(msg) {
		return true
	}

	var metaMsg = &mqMessage{}
	if err := utils.JsonToStruct(msg, metaMsg); err != nil {
		mongoutils.Error(err)
		return false
	}

	if utils.IsEmpty(metaMsg.JsonContent) {
		return true
	}

	result := consumer.Consume(metaMsg.JsonContent)
	if !result {
		// todo 重试
	} else {
		// todo 成功处理
	}

	return result
}
