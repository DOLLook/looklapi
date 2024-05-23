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
type grabSchedulerTask interface {
	// task executor
	executor() func()
	// use a redis key as hold flag
	redisHoldKey() string
	// redis lock key
	redisLockerKey() string
	// the key hold max seconds. the hold key will be released by redis automatically when timeout.
	redisKeyHoldSecs() int
	// start the task
	startTask(execWrapper func())
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

	for _, task := range wireutils.ResovleAll(reflect.TypeOf((*grabSchedulerTask)(nil)).Elem()) {
		if task, ok := task.(grabSchedulerTask); ok {
			task.startTask(manager.wrapper(task))
		}
	}
	manager.init = true
}

func (manager *grabSchedulerTaskManager) wrapper(task grabSchedulerTask) func() {
	return func() {
		defer loggers.RecoverLog()
		grab := manager.grab(task)
		if grab {
			task.executor()()
		}
	}
}

// grab the hold key
func (manager *grabSchedulerTaskManager) grab(task grabSchedulerTask) bool {
	doing := false
	if _, err := redisutils.LockAction(func() error {
		if exist, _ := redisutils.Exist(task.redisHoldKey()); exist {
			doing = true
			return nil
		}
		return redisutils.SetEx(task.redisHoldKey(), 1, task.redisKeyHoldSecs())

	}, task.redisLockerKey(), 10); err != nil {
		return false
	}

	return !doing
}

// manual release hold key
func (manager *grabSchedulerTaskManager) release(grab bool, task grabSchedulerTask) {
	if grab {
		if err := redisutils.Del(task.redisHoldKey()); err != nil {
			loggers.GetLogger().Warn(fmt.Sprintf("holdkey:%s realse failed", task.redisHoldKey()))
			loggers.GetLogger().Error(err)
		}
	}
}
