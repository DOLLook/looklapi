package mqutils

import (
	"errors"
	"github.com/streadway/amqp"
	"go-webapi-fw/common/loggers"
	serviceDiscovery "go-webapi-fw/common/service-discovery"
	"go-webapi-fw/common/utils"
	"go-webapi-fw/errs"
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
		panic(errs.NewBllError("invalid nil consumer"))
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
		panic(errs.NewBllError("invalid nil consumer"))
	}

	if consumer.Type != _WorkQueue {
		panic(errs.NewBllError("invalid consumer type"))
	}

	if consumer.Consume == nil {
		panic(errs.NewBllError("must bind consume func"))
	}

	recChan := getConsumerChannel()
	if _, err := recChan.Channel.QueueDeclare(consumer.RouteKey, true, false, false, false, nil); err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	if err := recChan.Channel.Qos(int(consumer.PrefetchCount), 0, false); err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	deliverCh, err := recChan.Channel.Consume(consumer.RouteKey, "", false, false, false, false, nil)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		for {
			select {
			case delivery := <-deliverCh:
				if consumer.Parallel && consumer.PrefetchCount > 1 {
					go consume(&delivery, consumer)
				} else {
					consume(&delivery, consumer)
				}
			case err := <-errChan:
				loggers.GetLogger().Error(err)
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
		panic(errs.NewBllError("invalid nil consumer"))
	}

	if consumer.Type != _Broadcast {
		panic(errs.NewBllError("invalid consumer type"))
	}

	if consumer.Consume == nil {
		panic(errs.NewBllError("must bind consume func"))
	}

	recChan := getConsumerChannel()
	if err := recChan.Channel.ExchangeDeclare(consumer.Exchange, "fanout", false, true, false, false, nil); err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	queue, err := recChan.Channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	if err := recChan.Channel.QueueBind(queue.Name, "", consumer.Exchange, false, nil); err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	deliverCh, err := recChan.Channel.Consume(queue.Name, "", false, true, false, false, nil)
	if err != nil {
		panic(errs.NewBllError(err.Error()))
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		for {
			select {
			case delivery := <-deliverCh:
				consume(&delivery, consumer)
			case err := <-errChan:
				loggers.GetLogger().Error(err)
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
			loggers.GetLogger().Error(errors.New("消息Nack失败: " + content))
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
		panic(errs.NewBllError("invalid routekey"))
	}

	if concurrency < 1 {
		panic(errs.NewBllError("workqueue consumer concurrency must greater than 0"))
	}

	if prefetchCount < 1 {
		panic(errs.NewBllError("workqueue consumer prefetchCount must greater than 0"))
	}

	if maxRetry < 1 {
		panic(errs.NewBllError("workqueue consumer maxRetry must greater than 0"))
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
		panic(errs.NewBllError("invalid exchange"))
	}

	if maxRetry < 1 {
		panic(errs.NewBllError("broadcast consumer maxRetry must greater than 0"))
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

	defer func() {
		if err := recover(); err != nil {
			if throws, ok := err.(error); ok {
				loggers.GetLogger().Error(throws)
			}
		}
	}()

	if consumer.Type == _WorkQueue {
		if serviceDiscovery.GetServiceManager().IsHostCutoff() {
			return false
		}
	}

	if utils.IsEmpty(msg) {
		return true
	}

	var metaMsg = &mqMessage{}
	if err := utils.JsonToStruct(msg, metaMsg); err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	if utils.IsEmpty(metaMsg.JsonContent) {
		return true
	}

	result := consumer.Consume(metaMsg.JsonContent)
	if !result {
		result = retry(metaMsg, consumer)
	} else {
		go retrySuccess(metaMsg, consumer.Type)
	}

	return result
}
