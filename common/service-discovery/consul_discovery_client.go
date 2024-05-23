package service_discovery

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"looklapi/common/loggers"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"strconv"
	"strings"
)

// consul注册信息
type consulServiceRegistry struct {
	client       *consulApi.Client
	registration *consulApi.AgentServiceRegistration
	config       *consulApi.Config
}

var serviceRegistry *consulServiceRegistry

// 注册服务
func register() {
	serviceRegistry = new(consulServiceRegistry)
	serviceRegistry.registration = generateAgentServiceRegistration()
	serviceRegistry.config = generateConsulConfig()

	if consulClient, err := consulApi.NewClient(serviceRegistry.config); err == nil {
		serviceRegistry.client = consulClient
	} else {
		loggers.GetLogger().Error(err)
		panic("consul client init failed")
	}

	// 注册服务到consul
	if err := serviceRegistry.client.Agent().ServiceRegister(serviceRegistry.registration); err != nil {
		loggers.GetLogger().Error(err)
		panic("consul register failed")
	}

	loggers.GetLogger().Info("consul register success")
}

// 获取所有健康服务
func getAllHealthServices() (map[string][]*consulApi.AgentService, error) {
	serviceMap, err := serviceRegistry.client.Agent().Services()
	if err != nil {
		return nil, err
	}

	var result = make(map[string][]*consulApi.AgentService)
	hcks, _, hckErr := serviceRegistry.client.Health().State(consulApi.HealthPassing, nil)
	if hckErr != nil {
		return nil, err
	}

	for _, hck := range hcks {
		if service, ok := serviceMap[hck.ServiceID]; ok {
			result[service.Service] = append(result[service.Service], service)
		}
	}

	return result, nil
}

func generateAgentServiceRegistration() *consulApi.AgentServiceRegistration {
	registration := new(consulApi.AgentServiceRegistration)
	registration.ID = fmt.Sprintf("%s-%s-%s",
		appConfig.AppConfig.Server.Name,
		strings.ReplaceAll(utils.HostIp(), ".", "-"),
		appConfig.AppConfig.Server.Port)
	registration.Name = appConfig.AppConfig.Server.Name
	registration.Address = utils.HostIp()
	registration.Port, _ = strconv.Atoi(appConfig.AppConfig.Server.Port)

	var tags []string
	if appConfig.AppConfig.Consul.Secure {
		tags = append(tags, "secure=true")
	} else {
		tags = append(tags, "secure=false")
	}
	registration.Tags = tags

	registration.Check = generateHealthCheck(registration.Address, registration.Port)

	return registration
}

func generateHealthCheck(address string, port int) *consulApi.AgentServiceCheck {
	check := new(consulApi.AgentServiceCheck)
	schema := "http"
	if appConfig.AppConfig.Consul.Secure {
		schema = "https"
	}
	check.HTTP = fmt.Sprintf("%s://%s:%d/%s",
		schema, address, port,
		strings.TrimLeft(appConfig.AppConfig.Consul.HealthCheck, "/"))

	check.Timeout = appConfig.AppConfig.Consul.Timeout
	check.Interval = appConfig.AppConfig.Consul.Interval
	check.DeregisterCriticalServiceAfter = appConfig.AppConfig.Consul.DeregisterCriticalServiceAfter // 故障检查失败后 consul自动将注册服务删除

	return check
}

func generateConsulConfig() *consulApi.Config {
	consulConfig := consulApi.DefaultConfig()
	consulConfig.Address = appConfig.AppConfig.Consul.Host + ":" + appConfig.AppConfig.Consul.Port
	consulConfig.Scheme = "http"

	return consulConfig
}
