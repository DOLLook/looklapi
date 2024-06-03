package loggers

import "github.com/pkg/errors"

func RecoverLog() {
	if err := recover(); err != nil {
		if throws, ok := err.(error); ok {
			if _, ok := throws.(stackTracer); ok {
				GetLogger().Error(throws)
			} else {
				GetLogger().Error(errors.WithStack(throws))
			}
		} else if msg, ok := err.(string); ok {
			GetLogger().Error(errors.New(msg))
		}
	}
}
