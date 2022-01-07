package loggers

import "errors"

func RecoverLog() {
	if err := recover(); err != nil {
		if throws, ok := err.(error); ok {
			GetLogger().Error(throws)
		} else if msg, ok := err.(string); ok {
			GetLogger().Error(errors.New(msg))
		}
	}
}
