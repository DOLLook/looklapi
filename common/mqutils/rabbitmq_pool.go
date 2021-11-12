package mqutils

import (
	"errors"
	"github.com/robfig/cron/v3"
	"github.com/streadway/amqp"
	"go-webapi-fw/common/loggers"
	"go-webapi-fw/common/utils"
	appConfig "go-webapi-fw/config"
	"go-webapi-fw/errs"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var _rabbitmqConnPool *rabbitmqConnPool // mq连接池
var _idlClearTask *cron.Cron            // 空闲连接清理任务

func init() {
	_rabbitmqConnPool = &rabbitmqConnPool{
		connStr: appConfig.AppConfig.Rabbitmq.Address,

		pubConns:      make(map[string]*rabbitMqConnData),
		pubChs:        make([]*mqChannel, 0, _chLimitForConn),
		pubChPipeline: make(chan *mqChannel, 1),
		//pubLock:       false,

		recConns: make([]*rabbitMqConnData, 0, _connLimt),
		recChs:   make([]*mqChannel, 0, _chLimitForConn),
		recMu:    &sync.Mutex{},
	}
	_rabbitmqConnPool.pubLock = new(int32)

	go func() {
		for {
			pushPubChToPipe()
		}
	}()

	startTimingTask()
}

// 添加消费者
func addConsumerChannel(consumer *mqChannel) {
	if consumer == nil || consumer.Channel == nil {
		return
	}

	_rabbitmqConnPool.recMu.Lock()
	defer _rabbitmqConnPool.recMu.Unlock()

	_rabbitmqConnPool.recChs = append(_rabbitmqConnPool.recChs, consumer)
	consumer.Conn.incChan()
}

// 删除消费者
func removeConsumerChannel(consumer *mqChannel) {
	if consumer == nil || consumer.Channel == nil {
		return
	}

	_rabbitmqConnPool.recMu.Lock()
	defer _rabbitmqConnPool.recMu.Unlock()

	if utils.SliceRemove(&_rabbitmqConnPool.recChs, consumer, 1) {
		//consumer.Conn.decChan()
		if consumer.Conn.Conn.IsClosed() {
			utils.SliceRemove(&_rabbitmqConnPool.recConns, consumer.Conn, 1)
		} else {
			consumer.Conn.decChan()
		}
	}
}

// 获取消费者通道
func getConsumerChannel() *mqChannel {
	commonsl := utils.NewCommonSlice(_rabbitmqConnPool.recConns)
	validConns := commonsl.Filter(func(item interface{}) bool {
		connModel := item.(*rabbitMqConnData)
		return *connModel.LiveCh < _chLimitForConn && !connModel.Conn.IsClosed()
	})

	var conn *rabbitMqConnData

	if validConns == nil || len(validConns) < 1 {
		if len(_rabbitmqConnPool.recConns) >= _connLimt {
			panic(errs.NewBllError("连接池已满"))
		}

		amqpConn, err := amqp.Dial(_rabbitmqConnPool.connStr)
		if err != nil {
			panic(errs.NewBllError(err.Error()))
		}

		conn = newRabbitMQConn(amqpConn)
		_rabbitmqConnPool.recConns = append(_rabbitmqConnPool.recConns, conn)

	} else {
		sort.Slice(validConns, func(i, j int) bool {
			vali := *validConns[i].(*rabbitMqConnData).LiveCh
			valj := *validConns[j].(*rabbitMqConnData).LiveCh
			return vali > valj
		})

		conn = validConns[0].(*rabbitMqConnData)
	}

	channel := newRabbitMQChannel(conn)
	if chn, err := conn.Conn.Channel(); err == nil {
		channel.Channel = chn
	} else {
		panic(errs.NewBllError(err.Error()))
	}

	return channel
}

/**
尝试获取有效的发布通道
maxTry 最大重试次数
*/
func tryGetPubChannel(maxTry int) (*mqChannel, error) {
	if maxTry <= 1 {
		return getPubChannel()
	}

	if pubChan, err := getPubChannel(); err != nil || pubChan.Status == _Close || pubChan.Conn.Conn.IsClosed() {
		return tryGetPubChannel(maxTry - 1)
	} else {
		return pubChan, nil
	}
}

// 获取发布通道
func getPubChannel() (*mqChannel, error) {
	pubChan := <-_rabbitmqConnPool.pubChPipeline
	if pubChan == nil {
		return nil, errors.New("连接池已满")
	}

	return pubChan, nil
}

// 向管道推送发布channel
func pushPubChToPipe() {
	for {
		if atomic.CompareAndSwapInt32(_rabbitmqConnPool.pubLock, 0, 1) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 查找空闲通道
	var idleCh *mqChannel
	for _, item := range _rabbitmqConnPool.pubChs {
		if item.Status != _Busy && item.Status != _Close && !item.Conn.Conn.IsClosed() {
			idleCh = item
			break
		}
	}

	if idleCh != nil {
		idleCh.Status = _Busy
		idleCh.updateLastUseMills()
		atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0)
		_rabbitmqConnPool.pubChPipeline <- idleCh
		return
	}

	// 未找到空闲通道 创建新通道
	connSlice := make([]*rabbitMqConnData, len(_rabbitmqConnPool.pubConns))
	connIndex := 0
	for _, conn := range _rabbitmqConnPool.pubConns {
		connSlice[connIndex] = conn
		connIndex++
	}
	commonsl := utils.NewCommonSlice(connSlice)
	validConns := commonsl.Filter(func(item interface{}) bool {
		connModel := item.(*rabbitMqConnData)
		return *connModel.LiveCh < _chLimitForConn && !connModel.Conn.IsClosed()
	})

	var conn *rabbitMqConnData

	if validConns == nil || len(validConns) < 1 {
		if len(_rabbitmqConnPool.pubConns) >= _connLimt {
			// 连接池已满
			atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0)
			_rabbitmqConnPool.pubChPipeline <- nil
			return
		}

		amqpConn, err := amqp.Dial(_rabbitmqConnPool.connStr)
		if err != nil {
			atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0)
			loggers.GetLogger().Error(err)
			_rabbitmqConnPool.pubChPipeline <- nil
			return
		}

		conn = newRabbitMQConn(amqpConn)
		_rabbitmqConnPool.pubConns[conn.Guid] = conn

	} else {
		sort.Slice(validConns, func(i, j int) bool {
			vali := *validConns[i].(*rabbitMqConnData).LiveCh
			valj := *validConns[j].(*rabbitMqConnData).LiveCh
			return vali > valj
		})

		conn = validConns[0].(*rabbitMqConnData)
	}

	channel := newRabbitMQChannel(conn)
	if chn, err := conn.Conn.Channel(); err == nil {
		channel.Channel = chn
	} else {
		atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0)
		loggers.GetLogger().Error(err)
		_rabbitmqConnPool.pubChPipeline <- nil
		return
	}

	_rabbitmqConnPool.pubChs = append(_rabbitmqConnPool.pubChs, channel)
	channel.Conn.incChan()
	atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0)

	errChan := make(chan *amqp.Error, 1)
	errChan = channel.Channel.NotifyClose(errChan)
	go func() {
		select {
		case <-errChan:
			channel.Status = _Close
		}
	}()

	_rabbitmqConnPool.pubChPipeline <- channel
}

// 启动定时任务
func startTimingTask() {
	_idlClearTask = cron.New(cron.WithSeconds())
	_, err := _idlClearTask.AddFunc("0 */1 * * * ?", clearIdlPubConn)

	if err != nil {
		loggers.GetLogger().Error(err)
	}

	_idlClearTask.Start()
}

// 清理空闲发布连接
func clearIdlPubConn() {
	holdLock := false
	for i := 0; i < 3; i++ {
		if atomic.CompareAndSwapInt32(_rabbitmqConnPool.pubLock, 0, 1) {
			holdLock = true
			break
		}
		// 尝试抢占3次
		time.Sleep(1 * time.Second)
	}

	if !holdLock {
		return
	}

	defer atomic.StoreInt32(_rabbitmqConnPool.pubLock, 0) // 释放锁

	now := time.Now().UnixNano() / 1000000
	for i := len(_rabbitmqConnPool.pubChs) - 1; i >= 0; i-- {
		item := _rabbitmqConnPool.pubChs[i]
		if item.Status == _Busy && !item.Conn.Conn.IsClosed() {
			continue
		}

		if item.Conn.Conn.IsClosed() {
			delete(_rabbitmqConnPool.pubConns, item.Conn.Guid)
			utils.SliceRemoveByIndex(&_rabbitmqConnPool.pubChs, i)
			continue
		}

		doClear := false
		if item.Status == _Close {
			doClear = true
		} else if now-item.LastUseMills >= _chIdleTimeoutMin*60*1000 {
			doClear = true
			item.Status = _Timeout
		}

		if doClear {
			if *item.Conn.LiveCh <= 1 {
				// 关闭连接
				if now-item.Conn.LastUseMills >= _connIdleTimeoutMin*60*1000 {
					utils.SliceRemoveByIndex(&_rabbitmqConnPool.pubChs, i)
					item.Conn.decChan()
					delete(_rabbitmqConnPool.pubConns, item.Conn.Guid)

					if item.Status == _Timeout {
						if closeErr := item.Channel.Close(); closeErr != nil {
							loggers.GetLogger().Error(closeErr)
						}
					}
					if !item.Conn.Conn.IsClosed() {
						if closeErr := item.Conn.Conn.Close(); closeErr != nil {
							loggers.GetLogger().Error(closeErr)
						}
					}
				}
			} else {
				// 仅关闭通道
				utils.SliceRemoveByIndex(&_rabbitmqConnPool.pubChs, i)
				item.Conn.decChan()

				if item.Status == _Timeout {
					if closeErr := item.Channel.Close(); closeErr != nil {
						loggers.GetLogger().Error(closeErr)
					}
				}
			}
		}
	}
}
