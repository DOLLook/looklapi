package web

import (
	"micro-webapi/common/appcontext"
	"micro-webapi/common/loggers"
	"micro-webapi/common/redisutils"
	"micro-webapi/model/modelimpl"
)

// load the remote loglevel config
func LoadLogConfig() {
	logConfig := &modelimpl.ConfigLog{}
	if err := redisutils.Get(redisutils.CONFIG_LOG, logConfig); err != nil {
		loggers.GetLogger().Error(err)
	} else {
		appcontext.GetAppEventPublisher().PublishEvent(logConfig)
	}
}
