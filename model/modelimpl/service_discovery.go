package modelimpl

// 手动服务配置信息
type ManualService struct {
	Cutoff  []string        // 断流服务
	Healthy []*ServiceModel // 健康服务
}

type ServiceModel struct {
	ServiceName string   // 服务名称
	Endpoints   []string // 终结点
}
