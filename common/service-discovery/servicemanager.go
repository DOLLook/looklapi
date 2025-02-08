package service_discovery

import (
	"fmt"
	"looklapi/common/appcontext"
	"looklapi/common/loggers"
	"looklapi/common/redisutils"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"looklapi/model/modelimpl"
	"reflect"
	"slices"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/robfig/cron/v3"
)

var svManager *serviceManager

// 服务管理者
type serviceManager struct {
	forceHealthyCache []*modelimpl.ServiceModel          // 强制保持正常服务缓存
	cutoffCache       []string                           // 强制断流服务缓存
	healthServices    map[string]*modelimpl.ServiceModel // 健康服务
	hostEndpoint      string                             // 本机endpoint
	updateTask        *cron.Cron                         // 服务更新任务
	isReady           bool                               // 是否就绪
	isCutoff          bool                               // 本机是否断流
}

func init() {
	svManager = &serviceManager{
		hostEndpoint: fmt.Sprintf("%s:%s", utils.HostIp(), appConfig.AppConfig.Server.Port),
		updateTask:   cron.New(cron.WithSeconds()),
	}
	svManager.Subscribe()
}

// received app event and process.
// for event publish well, the developers must deal with the panic by their self
func (manager *serviceManager) OnApplicationEvent(event interface{}) {
	manager.initialize()
}

// register to the application event publisher
func (manager *serviceManager) Subscribe() {
	appcontext.GetAppEventPublisher().Subscribe(manager, reflect.TypeOf(appcontext.AppEventBeanInjected(0)))
}

// 获取服务管理器
func GetServiceManager() *serviceManager {
	return svManager
}

// 初始服务化管理器
func (manager *serviceManager) initialize() {
	if manager.isReady {
		return
	}

	if utils.IsEmpty(appConfig.AppConfig.Consul.Host) || utils.IsEmpty(appConfig.AppConfig.Consul.Port) {
		// service discovery is not enabled
		manager.isReady = true
		return
	}

	register()

	manualService := &modelimpl.ManualService{}
	if err := redisutils.Get(redisutils.CONFIG_MANUAL_SERVICE, manualService); err == nil {
		if !utils.CollectionIsEmpty(manualService.Cutoff) {
			manager.cutoffCache = append(manager.cutoffCache, manualService.Cutoff...)
			manager.isCutoff = slices.Contains(manager.cutoffCache, manager.hostEndpoint)
		}

		if !utils.CollectionIsEmpty(manualService.Healthy) {
			manager.forceHealthyCache = append(manager.forceHealthyCache, manualService.Healthy...)
		}
	} else {
		loggers.GetLogger().Error(err)
	}

	_, err := manager.updateTask.AddFunc("*/10 * * * * ?", manager.updateHealthServices)
	if err != nil {
		loggers.GetLogger().Error(err)
	}
	manager.updateTask.Start()
	manager.isReady = true
}

// 更新服务配置
func (manager *serviceManager) UpdateManualService(manualService *modelimpl.ManualService) {
	cutoff := make([]string, 0)
	if manualService != nil && !utils.CollectionIsEmpty(manualService.Cutoff) {
		cutoff = append(cutoff, manualService.Cutoff...)
		manager.cutoffCache = cutoff
		manager.isCutoff = slices.Contains(cutoff, manager.hostEndpoint)
	} else {
		manager.cutoffCache = cutoff
		manager.isCutoff = false
	}

	forceHealthy := make([]*modelimpl.ServiceModel, 0)
	if manualService != nil && !utils.CollectionIsEmpty(manualService.Healthy) {
		forceHealthy = append(forceHealthy, manualService.Healthy...)
	}
	manager.forceHealthyCache = forceHealthy
}

// 更新服务自动发现健康服务
func (manager *serviceManager) updateHealthServices() {
	defer loggers.RecoverLog()

	hsrv, err := getAllHealthServices()
	if err != nil {
		loggers.GetLogger().Error(err)
		return
	}
	manager.healthServices = convertServiceModel(hsrv)
}

// 获取健康的服务
// serviceName 服务名称
func (manager *serviceManager) GetHealthServices(serviceName string) []string {
	var serviceList []string

	var hsCopy = manager.healthServices
	if model, ok := hsCopy[serviceName]; ok {
		endPoints := utils.NewCommonSlice(model.Endpoints)
		for _, ep := range endPoints.Filter(func(ep interface{}) bool {
			return !utils.ArrayOrSliceContains(manager.cutoffCache, ep)
		}) {
			serviceList = append(serviceList, ep.(string))
		}
	}

	if len(serviceList) < 1 {
		forceHealthyCopy := manager.forceHealthyCache
		forceHealthy := utils.NewCommonSlice(forceHealthyCopy).Filter(func(item interface{}) bool {
			return serviceName == item.(*modelimpl.ServiceModel).ServiceName
		})

		if len(forceHealthy) < 1 {
			return serviceList
		}

		serviceList = append(serviceList, forceHealthy[0].(modelimpl.ServiceModel).Endpoints...)
	}

	return serviceList
}

// 本机是否断流
func (manager *serviceManager) IsHostCutoff() bool {
	if !manager.isReady {
		return false
	}
	return manager.isCutoff
}

func convertServiceModel(agentServiceMap map[string][]*consulApi.AgentService) map[string]*modelimpl.ServiceModel {
	result := make(map[string]*modelimpl.ServiceModel, len(agentServiceMap))
	for key, val := range agentServiceMap {
		result[key] = &modelimpl.ServiceModel{ServiceName: key}
		for _, agentService := range val {
			result[key].Endpoints = append(result[key].Endpoints,
				fmt.Sprintf("%s:%d", agentService.Address, agentService.Port))
		}
	}
	return result
}
