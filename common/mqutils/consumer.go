package mqutils

import (
	"errors"
	"fmt"
	"looklapi/common/appcontext"
	"looklapi/common/loggers"
	serviceDiscovery "looklapi/common/service-discovery"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"looklapi/errs"
	"reflect"
	"time"

	"github.com/streadway/amqp"
)

type consumerBinder struct {
	workqueueReconnectCh chan *consumer
	broadcastReconnectCh chan *consumer
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (binder *consumerBinder) OnApplicationEvent(event interface{}) {
	defer loggers.RecoverLog()

	if _hasConsumerBind || _consumerContainer == nil || len(_consumerContainer) < 1 {
		return
	}

	workerCount, broadcasterCount := 0, 0
	for _, consumer := range _consumerContainer {
		switch consumer.Type {
		case _WorkQueue:
			workerCount += int(consumer.Concurrency)
		case _Broadcast:
			broadcasterCount += 1
		}
	}
	binder.workqueueReconnectCh = make(chan *consumer, workerCount)
	binder.broadcastReconnectCh = make(chan *consumer, broadcasterCount)

	go func() {
		defer loggers.RecoverLog()
		lastReconnectOk := true
		continueErr := 0
		for c := range binder.workqueueReconnectCh {
			if !lastReconnectOk {
				if continueErr > 10 {
					continueErr = 10
				}
				waitSecs := continueErr
				if waitSecs < 1 {
					waitSecs = 1
				}
				loggers.GetLogger().Warn(fmt.Sprintf("last reconnect failed, wait %d seconds to reconnect", waitSecs))
				time.Sleep(time.Duration(waitSecs) * time.Second)
			}

			lastReconnectOk = binder.bindWorkQueueConsumer(c)
			if lastReconnectOk {
				continueErr = 0
				loggers.GetLogger().Warn("reconnect success, workqueue consumer: " + c.RouteKey)
			} else {
				continueErr += 1
				binder.workqueueReconnectCh <- c
			}
		}
	}()

	go func() {
		defer loggers.RecoverLog()
		lastReconnectOk := true
		continueErr := 0
		for c := range binder.broadcastReconnectCh {
			if !lastReconnectOk {
				if continueErr > 10 {
					continueErr = 10
				}
				waitSecs := continueErr
				if waitSecs < 1 {
					waitSecs = 1
				}
				loggers.GetLogger().Warn(fmt.Sprintf("last reconnect failed, wait %d seconds to reconnect", waitSecs))
				time.Sleep(time.Duration(waitSecs) * time.Second)
			}

			lastReconnectOk = binder.bindBroadcastConsumer(c)
			if lastReconnectOk {
				continueErr = 0
				loggers.GetLogger().Warn("reconnect success, broadcast consumer: " + c.Exchange)
			} else {
				continueErr += 1
				binder.broadcastReconnectCh <- c
			}
		}
	}()

	for _, item := range _consumerContainer {
		binder.bindConsumer(item)
	}
	_hasConsumerBind = true
	loggers.GetLogger().Info("mq init complete")
}

// register to the application event publisher
func (binder *consumerBinder) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(binder, reflect.TypeOf(appcontext.AppEventBeanInjected(0)))
}

func init() {
	if utils.IsEmpty(appConfig.AppConfig.Rabbitmq.Address) {
		return
	}
	binder := &consumerBinder{}
	binder.Subscribe()
}

// 绑定消费者
func (binder *consumerBinder) bindConsumer(consumer *consumer) {
	if consumer == nil {
		loggers.GetLogger().Error(errors.New("invalid nil consumer"))
		return
	}

	switch consumer.Type {
	case _WorkQueue:
		for i := uint32(0); i < consumer.Concurrency; i++ {
			if !binder.bindWorkQueueConsumer(consumer) {
				binder.workqueueReconnectCh <- consumer
			}
		}
	case _Broadcast:
		if !binder.bindBroadcastConsumer(consumer) {
			binder.broadcastReconnectCh <- consumer
		}
	}
}

// 绑定工作队列消费者
// consumer 消费者
func (binder *consumerBinder) bindWorkQueueConsumer(consumer *consumer) bool {
	if consumer == nil {
		loggers.GetLogger().Error(errors.New("invalid nil consumer"))
		return false
	}

	if consumer.Type != _WorkQueue {
		loggers.GetLogger().Error(errors.New("invalid consumer type"))
		return false
	}

	if consumer.Consume == nil {
		loggers.GetLogger().Error(errors.New("must bind consume func"))
		return false
	}

	recChan, err := getConsumerChannel()
	if err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	if _, err := recChan.Channel.QueueDeclare(consumer.RouteKey, true, false, false, false, nil); err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	if err := recChan.Channel.Qos(int(consumer.PrefetchCount), 0, false); err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	deliverCh, err := recChan.Channel.Consume(consumer.RouteKey, "", false, false, false, false, nil)
	if err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		defer loggers.RecoverLog()
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
				// bindWorkQueueConsumer(consumer)
				binder.workqueueReconnectCh <- consumer
				return
			}
		}
	}()

	addConsumerChannel(recChan)
	return true
}

// 绑定广播队列消费者
// consumer 消费者
func (binder *consumerBinder) bindBroadcastConsumer(consumer *consumer) bool {
	if consumer == nil {
		loggers.GetLogger().Error(errors.New("invalid nil consumer"))
		return false
	}

	if consumer.Type != _Broadcast {
		loggers.GetLogger().Error(errors.New("invalid consumer type"))
		return false
	}

	if consumer.Consume == nil {
		loggers.GetLogger().Error(errors.New("must bind consume func"))
		return false
	}

	recChan, err := getConsumerChannel()
	if err != nil {
		loggers.GetLogger().Error(err)
		return false
	}
	if err := recChan.Channel.ExchangeDeclare(consumer.Exchange, "fanout", false, true, false, false, nil); err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	queue, err := recChan.Channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	if err := recChan.Channel.QueueBind(queue.Name, "", consumer.Exchange, false, nil); err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	deliverCh, err := recChan.Channel.Consume(queue.Name, "", false, true, false, false, nil)
	if err != nil {
		loggers.GetLogger().Error(err)
		return false
	}

	errChan := make(chan *amqp.Error, 1)
	errChan = recChan.Channel.NotifyClose(errChan)

	go func() {
		defer loggers.RecoverLog()
		for {
			select {
			case delivery := <-deliverCh:
				consume(&delivery, consumer)
			case err := <-errChan:
				loggers.GetLogger().Error(err)
				removeConsumerChannel(recChan)
				// bindBroadcastConsumer(consumer)
				binder.broadcastReconnectCh <- consumer
				return
			}
		}
	}()

	addConsumerChannel(recChan)
	return true
}

func consume(delivery *amqp.Delivery, consumer *consumer) {
	content := string(delivery.Body)
	if consumer.onReceived(content) {
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
	Type        consumerType               // 消费者类型
	MaxRetry    uint32                     // 最大重试次数
	MessageType reflect.Type               // 消息类型
	Consume     func(msg interface{}) bool // 处理器

	Exchange string // broadcast交换器名称

	RouteKey      string // workqueue路由地址
	Concurrency   uint32 // workqueue并发消费者数量
	PrefetchCount uint32 // workqueue从队列中同时deliver的消息数量
	Parallel      bool   // workqueue是否开启并行消费
}

// 新建消费者
func NewWorkQueueConsumer(routeKey string, concurrency uint32, prefetchCount uint32, parallel bool, maxRetry uint32, messageType reflect.Type, consume func(msg interface{}) bool) {
	if utils.IsEmpty(routeKey) {
		err := errs.NewBllError("invalid routekey")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	if concurrency < 1 {
		err := errs.NewBllError("workqueue consumer concurrency must greater than 0")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	if prefetchCount < 1 {
		err := errs.NewBllError("workqueue consumer prefetchCount must greater than 0")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	if maxRetry < 1 {
		err := errs.NewBllError("workqueue consumer maxRetry must greater than 0")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	cs := &consumer{
		Type:          _WorkQueue,
		MaxRetry:      maxRetry,
		RouteKey:      routeKey,
		Concurrency:   concurrency,
		PrefetchCount: prefetchCount,
		Parallel:      parallel,
		MessageType:   messageType,
		Consume:       consume,
	}

	_consumerContainer = append(_consumerContainer, cs)
}

// 新建消费者
func NewBroadcastConsumer(exchange string, maxRetry uint32, messageType reflect.Type, consume func(msg interface{}) bool) {
	if utils.IsEmpty(exchange) {
		err := errs.NewBllError("invalid exchange")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	if maxRetry < 1 {
		err := errs.NewBllError("broadcast consumer maxRetry must greater than 0")
		loggers.GetConsoleLogger().Error(err)
		panic(err)
	}

	cs := &consumer{
		Type:        _Broadcast,
		MaxRetry:    maxRetry,
		Exchange:    exchange,
		MessageType: messageType,
		Consume:     consume,
	}

	_consumerContainer = append(_consumerContainer, cs)
}

// 接收到消息
func (consumer *consumer) onReceived(msg string) (result bool) {
	defer func() {
		if err := recover(); err != nil {
			if tr, ok := err.(error); ok {
				loggers.GetLogger().Error(tr)
			} else if msg, ok := err.(string); ok {
				loggers.GetLogger().Error(errors.New(msg))
			}

			// 返回false 以便连续错误日志发现错误 若未成功写入重试信息可能导致无限重入 sleep 2sec 降低重入压力
			time.Sleep(2000 * time.Millisecond)
			result = false
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
		return true
	}

	if utils.IsEmpty(metaMsg.JsonContent) {
		return true
	}

	isptr := false
	tp := consumer.MessageType
	if tp.Kind() == reflect.Ptr {
		isptr = true
		tp = tp.Elem()
	}
	ptr := reflect.New(tp)
	if err := utils.JsonToStruct(metaMsg.JsonContent, ptr.Interface()); err != nil {
		loggers.GetLogger().Error(err)
		return retry(metaMsg, consumer)
	}

	msgobj := ptr.Interface()
	if !isptr {
		msgobj = ptr.Elem().Interface()
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				if tr, ok := err.(error); ok {
					loggers.GetLogger().Error(tr)
				} else if msg, ok := err.(string); ok {
					loggers.GetLogger().Error(errors.New(msg))
				}

				result = false
			}
		}()

		result = consumer.Consume(msgobj)
	}()

	if !result {
		result = retry(metaMsg, consumer)
	} else {
		go retrySuccess(metaMsg, consumer.Type)
	}

	return result
}
