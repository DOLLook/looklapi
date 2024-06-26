package mqutils

import (
	"github.com/streadway/amqp"
	"looklapi/common/loggers"
	"looklapi/common/utils"
	appConfig "looklapi/config"
)

// 发布工作队列消息
func PubWorkQueueMsg(routeKey string, msg interface{}) bool {
	if utils.IsEmpty(appConfig.AppConfig.Rabbitmq.Address) {
		loggers.GetLogger().Warn("mq is not enabled")
		return false
	}

	if utils.IsEmpty(routeKey) {
		return false
	}

	metaMsg := convertMessage(msg)
	if metaMsg == nil {
		return false
	}

	jsonMsg := utils.StructToJson(metaMsg)
	if ok, err := pubWorkQueueMsg(routeKey, jsonMsg); !ok {
		loggers.GetLogger().Error(err)
		return false
	}

	return true
}

// 发布广播消息
func PubBroadcastMsg(exchange string, msg interface{}) bool {
	if utils.IsEmpty(appConfig.AppConfig.Rabbitmq.Address) {
		loggers.GetLogger().Warn("mq is not enabled")
		return false
	}
	if utils.IsEmpty(exchange) {
		return false
	}

	metaMsg := convertMessage(msg)
	if metaMsg == nil {
		return false
	}

	jsonMsg := utils.StructToJson(metaMsg)
	if ok, err := pubBroadcastMsg(exchange, jsonMsg); !ok {
		loggers.GetLogger().Error(err)
		return false
	}

	return true
}

// 发布工作队列消息
func pubWorkQueueMsg(routeKey string, jsonMsg string) (bool, error) {
	var pubChan *mqChannel
	pubChan, err := tryGetPubChannel(3)
	if err != nil {
		return false, err
	}

	if _, err := pubChan.Channel.QueueDeclare(routeKey, true, false, false, false, nil); err != nil {
		return false, err
	}

	pubErr := pubChan.Channel.Publish("", routeKey, false, false,
		amqp.Publishing{
			//ContentType:  "text/plain",
			ContentType:  "application/octet-stream",
			Body:         []byte(jsonMsg),
			DeliveryMode: amqp.Persistent,
		})

	pubChan.Status = _Idle
	pubChan.updateLastUseMills()
	if pubErr != nil {
		return false, pubErr
	}

	return true, nil
}

// 发布广播消息
func pubBroadcastMsg(exchange string, jsonMsg string) (bool, error) {
	var pubChan *mqChannel
	pubChan, err := tryGetPubChannel(3)
	if err != nil {
		return false, err
	}

	if err := pubChan.Channel.ExchangeDeclare(exchange, "fanout", false, true, false, false, nil); err != nil {
		return false, err
	}

	pubErr := pubChan.Channel.Publish(exchange, "", false, false,
		amqp.Publishing{
			//ContentType:  "text/plain",
			ContentType: "application/octet-stream",
			Body:        []byte(jsonMsg),
		})

	pubChan.Status = _Idle
	pubChan.updateLastUseMills()
	if pubErr != nil {
		return false, pubErr
	}

	return true, nil
}

func convertMessage(msg interface{}) *mqMessage {
	if msg == nil {
		return nil
	}

	var metaMsg *mqMessage
	if tempMsg, ok := msg.(*mqMessage); !ok {
		metaMsg = newMessage()
		metaMsg.JsonContent = utils.StructToJson(msg)
	} else {
		metaMsg = tempMsg
	}

	if utils.IsEmpty(metaMsg.JsonContent) {
		return nil
	}

	return metaMsg
}
