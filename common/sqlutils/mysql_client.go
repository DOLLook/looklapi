package sqlutils

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	appConfig "looklapi/config"
	"strings"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var mysqlConf = appConfig.AppConfig.MySql
var xormMySqlEngine *xorm.Engine

func init() {
	if len(strings.TrimSpace(mysqlConf)) <= 0 {
		return
	}
	engine, err := xorm.NewEngine("mysql", mysqlConf)
	if err != nil {
		panic(fmt.Sprintf("can not init mysql engine. err:%s", err.Error()))
	}
	engine.SetMaxOpenConns(4)
	engine.SetMaxIdleConns(2)
	engine.SetColumnMapper(names.SnakeMapper{})
	engine.SetTableMapper(names.SnakeMapper{})
	if appConfig.AppConfig.Profile == "dev" {
		engine.ShowSQL(true)
	}
	xormMySqlEngine = engine
}

// 获取Mysql数据库引擎
func GetMySqlEngine() *xorm.Engine {
	return xormMySqlEngine
}
