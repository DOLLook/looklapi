package task

import "looklapi/common/loggers"

// 异步执行
func GoRunTask(f func(), callback func()) {
	go func() {
		defer loggers.RecoverLog()
		f()
		if callback != nil {
			callback()
		}
	}()
}
