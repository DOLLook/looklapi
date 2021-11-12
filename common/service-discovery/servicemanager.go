package service_discovery

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"go-webapi-fw/common/mongoutils"
	"go-webapi-fw/common/redisutils"
	"go-webapi-fw/common/utils"
	appConfig "go-webapi-fw/config"
	"go-webapi-fw/model/modelimpl"
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
}

func init() {
	svManager = &serviceManager{
		hostEndpoint: fmt.Sprintf("%s:%s", utils.HostIp(), appConfig.AppConfig.Server.Port),
		updateTask:   cron.New(cron.WithSeconds()),
	}
}

/**
获取服务管理器
*/
func GetServiceManager() *serviceManager {
	return svManager
}

/**
初始服务化管理器
*/
func (manager *serviceManager) Init() {
	if manager.isReady {
		return
	}

	register()

	manualService := &modelimpl.ManualService{}
	redisutils.Get(redisutils.CONFIG_MANUAL_SERVICE, manualService)

	if !utils.CollectionIsEmpty(manualService.Cutoff) {
		manager.cutoffCache = append(svManager.cutoffCache, manualService.Cutoff...)
	}

	if !utils.CollectionIsEmpty(manualService.Healthy) {
		manager.forceHealthyCache = append(svManager.forceHealthyCache, manualService.Healthy...)
	}

	_, err := manager.updateTask.AddFunc("*/10 * * * * ?", manager.updateHealthServices)
	if err != nil {
		mongoutils.Error(err)
	}
	manager.updateTask.Start()
	manager.isReady = true
}

/**
更新手动服务配置
*/
func (manager *serviceManager) UpdateManualService(manualService *modelimpl.ManualService) {
	manager.cutoffCache = manager.cutoffCache[0:0]
	if manualService != nil && !utils.CollectionIsEmpty(manualService.Cutoff) {
		manager.cutoffCache = append(manager.cutoffCache, manualService.Cutoff...)
	}

	manager.forceHealthyCache = manager.forceHealthyCache[0:0]
	if manualService != nil && !utils.CollectionIsEmpty(manualService.Healthy) {
		manager.forceHealthyCache = append(manager.forceHealthyCache, manualService.Healthy...)
	}
}

/**
更新服务自动发现健康服务
*/
func (manager *serviceManager) updateHealthServices() {
	hsrv, err := getAllHealthServices()
	if err != nil {
		mongoutils.Error(err)
		return
	}
	manager.healthServices = convertServiceModel(hsrv)
}

/**
获取健康的服务
serviceName 服务名称
*/
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
		forceHealthy := utils.NewCommonSlice(manager.forceHealthyCache).Filter(func(item interface{}) bool {
			return serviceName == item.(*modelimpl.ServiceModel).ServiceName
		})

		if len(forceHealthy) < 1 {
			panic(errors.New("找不到服务"))
		}

		serviceList = append(serviceList, forceHealthy[0].(modelimpl.ServiceModel).Endpoints...)
	}

	return serviceList
}

/**
本机是否断流
*/
func (manager *serviceManager) IsHostCutoff() bool {
	if !manager.isReady {
		return false
	}
	return utils.ArrayOrSliceContains(manager.cutoffCache, manager.hostEndpoint)
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
