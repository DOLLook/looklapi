package mqutils

import (
	"github.com/gofrs/uuid"
	"github.com/streadway/amqp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// connect 空闲超时时间 分钟
	_connIdleTimeoutMin = 10
	// channel 空闲超时时间 分钟
	_chIdleTimeoutMin = 5

	// connect 数量限制
	_connLimt = 100
	// channel 数量限制
	_chLimitForConn = 100
)

// 通道状态
type channelStatus byte

// 通道状态枚举
const (
	_Idle    channelStatus = iota // 空闲
	_Busy                         // 使用中
	_Timeout                      // 超时空闲
	_Close                        // 关闭
)

// rabbitmq连接池
type rabbitmqConnPool struct {
	connStr string // 连接地址

	pubConns      map[string]*rabbitMqConnData // 发布连接
	pubChs        []*mqChannel                 // 发布channel pool
	pubLock       *int32                       // 发布连接状态锁
	pubChPipeline chan *mqChannel              // 发布channel管道

	recConns []*rabbitMqConnData // 消费连接
	recChs   []*mqChannel        // 消费者channel pool
	recMu    *sync.Mutex         // 消费者管理锁
}

// rabbitmq连接数据
type rabbitMqConnData struct {
	Guid         string           // 连接id
	Conn         *amqp.Connection // 连接
	LiveCh       *int32           // 存活channel数
	LastUseMills int64            // 最近一次使用时间 毫秒
}

// 增加通道数
func (conn *rabbitMqConnData) incChan() {
	atomic.AddInt32(conn.LiveCh, 1)
	conn.LastUseMills = time.Now().UnixNano() / 1000000
}

// 减少通道
func (conn *rabbitMqConnData) decChan() {
	atomic.AddInt32(conn.LiveCh, -1)
}

// 新建连接模型
func newRabbitMQConn(conn *amqp.Connection) *rabbitMqConnData {
	uid, _ := uuid.NewV4()
	model := &rabbitMqConnData{
		Guid:         strings.ReplaceAll(uid.String(), "-", ""),
		Conn:         conn,
		LastUseMills: time.Now().UnixNano() / 1000000,
	}

	model.LiveCh = new(int32)
	return model
}

// MQ消息通道
type mqChannel struct {
	Conn         *rabbitMqConnData // 连接信息
	Channel      *amqp.Channel     // 通道
	Status       channelStatus     // 通道状态
	LastUseMills int64             // 最近一次使用时间 毫秒
}

// 新建MQ消息通道
func newRabbitMQChannel(conn *rabbitMqConnData) *mqChannel {
	now := time.Now().UnixNano() / 1000000
	conn.LastUseMills = now
	return &mqChannel{
		Conn:         conn,
		Status:       _Busy,
		LastUseMills: now,
	}
}

// 更新最近使用时间
func (ch *mqChannel) updateLastUseMills() {
	ch.LastUseMills = time.Now().UnixNano() / 1000000
	ch.Conn.LastUseMills = ch.LastUseMills
}

// 消息体
type mqMessage struct {
	Guid         string    // 消息id
	Timespan     time.Time `time_format:"SimpleDatetime"` // 消息生成时间
	CurrentRetry int32     // 当前重试次数
	JsonContent  string    // 消息内容
}

// 新建消息
func newMessage() *mqMessage {
	uid, _ := uuid.NewV4()
	return &mqMessage{
		Guid:         strings.ReplaceAll(uid.String(), "-", ""),
		Timespan:     time.Now(),
		CurrentRetry: 0,
		JsonContent:  "",
	}
}
