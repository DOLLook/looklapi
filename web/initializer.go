package web

import (
	"go-webapi-fw/common/appcontext"
	"go-webapi-fw/common/loggers"
	"go-webapi-fw/common/redisutils"
	"go-webapi-fw/model/modelimpl"
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
