// 配置文件包
package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// 应用程序配置
var AppConfig = &appConfig

// 程序配置
var appConfig = struct {
	Server struct {
		Name string `yaml:"name"`
		Port string `yaml:"port"`
	} `yaml:"server"`

	MySql      string `yaml:"mysql"`
	MongodbUri string `yaml:"mongodb_uri"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		// 超时时间 毫秒
		Timeout int32 `yaml:"timeout"`
	} `yaml:"redis"`

	Rabbitmq struct {
		Address string `yaml:"address"`
	} `yaml:"rabbitmq"`

	Consul struct {
		Host                           string `yaml:"host"`
		Port                           string `yaml:"port"`
		Secure                         bool   `yaml:"secure"`
		HealthCheck                    string `yaml:"health-check"`
		Timeout                        string `yaml:"timeout"`
		Interval                       string `yaml:"interval"`
		DeregisterCriticalServiceAfter string `yaml:"deregister-critical-service-after"`
	} `yaml:"consul"`

	Logger struct {
		Default string `yaml:"default"`
	} `yaml:"logger"`
}{}

func init() {
	bytes, err := ioutil.ReadFile("application.yml")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := yaml.Unmarshal(bytes, AppConfig); err != nil {
		fmt.Println(err)
	}
}
