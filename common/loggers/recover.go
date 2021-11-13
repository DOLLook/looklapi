package loggers

func RecoverLog() {
	if err := recover(); err != nil {
		if err, ok := err.(error); ok {
			GetLogger().Error(err)
		}
	}
}
