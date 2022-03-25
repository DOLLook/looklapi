package mqutils

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"looklapi/common/loggers"
	"looklapi/common/mongoutils"
	"looklapi/common/utils"
	"time"
)

const _RetryCollectionName = "mq_msg_retry" // 重试记录表

/**
重试消息
*/
func retry(metaMsg *mqMessage, consumer *consumer) bool {
	switch consumer.Type {
	case _WorkQueue:
		return retryWorker(metaMsg, consumer.RouteKey, consumer.MaxRetry)
	case _Broadcast:
		return retryBroadcast(metaMsg, consumer.Exchange, consumer.MaxRetry)
	default:
		loggers.GetLogger().Warn("unvalid consumer type")
		return false
	}
}

/**
重试工作队列消息
*/
func retryWorker(metaMsg *mqMessage, routeKey string, maxRetry uint32) bool {
	mqRetry := &mqMsgRetry{
		Id:          primitive.NewObjectID().Hex(),
		PublishType: int32(_WorkQueue),
		Key:         routeKey,
		Guid:        metaMsg.Guid,
		Timestamp:   metaMsg.Timespan,
		MaxRetry:    int32(maxRetry),
		Body:        metaMsg.JsonContent,
		Status:      0,
	}

	if metaMsg.CurrentRetry >= mqRetry.MaxRetry {
		// 达到最大重试次数
		addOrUpdate(mqRetry)
		return true
	} else {
		retryCount := getRetryCount(mqRetry.Guid, mqRetry.Timestamp, false)
		if retryCount != utils.MaxInt32 && retryCount >= mqRetry.MaxRetry {
			// 达到最大重试次数
			return true
		}
	}

	if mqRetry.Timestamp.UnixNano()/1000000+30*60*1000 <= time.Now().UnixNano()/1000000 {
		// 30分钟未成功消费的放弃掉
		errmsg := fmt.Sprintf("routekey:%s, guid:%s, timespan:%s 超过30分钟未消费",
			routeKey, mqRetry.Guid, mqRetry.Timestamp.Format("1970-01-01 00:00:00"))

		loggers.GetLogger().Warn(errmsg)
		return true
	}

	addOrUpdate(mqRetry)

	time.Sleep(1000 * time.Millisecond) // 等待1秒

	// 重新发布到队尾
	metaMsg.CurrentRetry += 1

	return PubWorkQueueMsg(routeKey, metaMsg)
}

/**
重试广播队列消息
*/
func retryBroadcast(metaMsg *mqMessage, exchange string, maxRetry uint32) bool {
	retrycount := getRetryCount(metaMsg.Guid, metaMsg.Timespan, true)
	if retrycount != utils.MaxInt32 && retrycount >= int32(maxRetry) {
		// 达到最大重试次数
		return true
	}

	if metaMsg.Timespan.UnixNano()/1000000+5*60*1000 <= time.Now().UnixNano()/1000000 {
		// 5分钟未成功消费的广播消息放弃掉
		errmsg := fmt.Sprintf("exchange:%s, guid:%s, timespan:%s 超过5分钟未消费",
			exchange, metaMsg.Guid, metaMsg.Timespan.Format("1970-01-01 00:00:00"))

		loggers.GetLogger().Warn(errmsg)
		return true
	}

	// 写入mongodb
	mqretry := &mqMsgRetry{
		PublishType: int32(_Broadcast),
		Key:         exchange,
		Guid:        metaMsg.Guid,
		Timestamp:   metaMsg.Timespan,
		MaxRetry:    int32(maxRetry),
		Body:        metaMsg.JsonContent,
		Status:      0,
	}

	result := false
	if !addOrUpdate(mqretry) {
		// 发生错误时返回true, 防止死循环
		result = true
		errmsg := fmt.Sprintf("消息重试失败, exchange:%s, guid:%s, timespan:%s.",
			exchange, metaMsg.Guid, metaMsg.Timespan.Format("1970-01-01 00:00:00"))
		loggers.GetLogger().Warn(errmsg)
	} else {
		// 等待两秒入队重试
		time.Sleep(2000 * time.Millisecond)
	}

	return result
}

/**
重试成功
*/
func retrySuccess(metaMsg *mqMessage, csType consumerType) {
	if metaMsg == nil || utils.IsEmpty(metaMsg.Guid) {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			if tr, ok := err.(error); ok {
				fmt.Println(tr.Error())
			} else if msg, ok := err.(string); ok {
				fmt.Println(msg)
			}
		}
	}()

	filter := bson.D{
		{"guid", metaMsg.Guid},
		{"timespan", metaMsg.Timespan},
	}

	if csType == _Broadcast {
		filter = append(filter, bson.E{Key: "hostIp", Value: utils.HostIp()})
	}

	updates := bson.D{
		{"$inc", bson.D{{"current_retry", 1}}},
		{"$set", bson.D{{"status", 1}, {"update_time", time.Now()}}},
	}

	if _, err := mongoutils.GetCollection(_RetryCollectionName).UpdateOne(nil, filter, updates); err != nil {
		fmt.Println(err.Error())
	}
}

/**
获取重试次数
*/
func getRetryCount(guid string, timespan time.Time, isBroadcast bool) int32 {

	if utils.IsEmpty(guid) {
		return utils.MaxInt32
	}

	result := int32(0)

	filter := bson.D{
		{"guid", guid},
		{"timespan", timespan},
	}

	if isBroadcast {
		filter = append(filter, bson.E{Key: "hostIp", Value: utils.HostIp()})
	}

	singleOne := mongoutils.GetCollection(_RetryCollectionName).FindOne(nil, filter)
	msgRetry := &mqMsgRetry{}

	if err := singleOne.Decode(msgRetry); err == nil {
		result = msgRetry.CurrentRetry
	} else {
		result = utils.MaxInt32
		loggers.GetLogger().Error(err)
	}

	return result
}

/**
新增或更新
*/
func addOrUpdate(model *mqMsgRetry) bool {
	if model == nil || utils.IsEmpty(model.Guid) || utils.IsEmpty(model.Key) {
		return false
	}

	filter := bson.D{
		{"guid", model.Guid},
		{"timespan", model.Timestamp},
	}

	if model.PublishType == int32(_Broadcast) {
		filter = append(filter, bson.E{Key: "hostIp", Value: utils.HostIp()})
	}

	collection := mongoutils.GetCollection(model.TbCollName())
	exist := false
	if count, err := collection.CountDocuments(nil, filter); err == nil && count > 0 {
		exist = true
	}

	result := false
	if !exist {
		// add
		if model.PublishType == int32(_Broadcast) {
			model.HostIp = utils.HostIp()
		} else {
			model.HostIp = ""
		}

		model.CurrentRetry = 0
		model.UpdateTime = time.Now()
		if inResult, err := collection.InsertOne(nil, model); err == nil {
			fmt.Println(inResult)
			result = true
		}
	} else {
		// update
		updates := bson.D{
			{"$inc", bson.D{{"current_retry", 1}}},
			{"$set", bson.D{{"update_time", time.Now()}}},
		}

		if updateResult, err := collection.UpdateOne(nil, filter, updates); err == nil {
			fmt.Println(updateResult)
			result = true
		}
	}

	return result
}
