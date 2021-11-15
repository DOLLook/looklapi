package mqutils

import (
	"time"
)

// mq重试记录
type mqMsgRetry struct {
	Id           string    `bson:"_id"`
	PublishType  int32     `bson:"publish_type"`  // 消息发布类型 consumerType
	Key          string    `bson:"key"`           // 消息路由
	Guid         string    `bson:"guid"`          // 消息id
	Timestamp    time.Time `bson:"timespan"`      // 消息生成时间
	CurrentRetry int32     `bson:"current_retry"` // 当前重试次数
	MaxRetry     int32     `bson:"max_retry"`     // 最大重试次数
	Body         string    `bson:"body"`          // 消息体
	UpdateTime   time.Time `bson:"update_time"`   // 更新时间
	Status       byte      `bson:"status"`        // 消费状态 0 未消费, 1 已消费
	HostIp       string    `bson:"hostIp"`        // 宿主ip
}

// 获取集合名称
func (log *mqMsgRetry) TbCollName() string {
	return _RetryCollectionName
}
