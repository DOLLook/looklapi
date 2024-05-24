package task

import (
	"fmt"
	"looklapi/common/appcontext"
	"looklapi/common/loggers"
	"looklapi/common/redisutils"
	"looklapi/common/wireutils"
	"reflect"
)

// grab task interface based on redis
type GrabSchedulerTask interface {
	// task executor
	Executor() func()
	// use a redis key as hold flag
	RedisHoldKey() string
	// redis lock key
	RedisLockerKey() string
	// the key hold max seconds. the hold key will be released by redis automatically when timeout.
	RedisKeyHoldSecs() int
	// start the task
	StartTask(execWrapper func())
}

// task manager
type grabSchedulerTaskManager struct {
	init bool
}

func init() {
	manager := &grabSchedulerTaskManager{}
	manager.Subscribe()
}

// register to the application event publisher
func (manager *grabSchedulerTaskManager) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(manager, reflect.TypeOf(appcontext.AppEventBeanInjected(0)))
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (manager *grabSchedulerTaskManager) OnApplicationEvent(event interface{}) {
	if manager.init {
		return
	}
	// 启动任务
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	for _, task := range wireutils.ResovleAll(reflect.TypeOf((*GrabSchedulerTask)(nil)).Elem()) {
		if task, ok := task.(GrabSchedulerTask); ok {
			task.StartTask(manager.wrapper(task))
		}
	}
	manager.init = true
}

func (manager *grabSchedulerTaskManager) wrapper(task GrabSchedulerTask) func() {
	return func() {
		defer loggers.RecoverLog()
		grab := manager.grab(task)
		if grab {
			task.Executor()()
		}
	}
}

// grab the hold key
func (manager *grabSchedulerTaskManager) grab(task GrabSchedulerTask) bool {
	doing := false
	if _, err := redisutils.LockAction(func() error {
		if exist, _ := redisutils.Exist(task.RedisHoldKey()); exist {
			doing = true
			return nil
		}
		return redisutils.SetEx(task.RedisHoldKey(), 1, task.RedisKeyHoldSecs())

	}, task.RedisLockerKey(), 10); err != nil {
		return false
	}

	return !doing
}

// manual release hold key
func (manager *grabSchedulerTaskManager) release(grab bool, task GrabSchedulerTask) {
	if grab {
		if err := redisutils.Del(task.RedisHoldKey()); err != nil {
			loggers.GetLogger().Warn(fmt.Sprintf("holdkey:%s realse failed", task.RedisHoldKey()))
			loggers.GetLogger().Error(err)
		}
	}
}
