package wireutils

import "reflect"

// 绑定映射模型
type wiredModel struct {
	metaType      reflect.Type  // 元类型
	priority      int           // 优先级
	proxy         bool          // 是否为代理类型
	target        interface{}   // 实例
	reflectTarget reflect.Value // 反射值
	injected      bool          // 是否注入完成
	injecting     bool          // 依赖注入中
}

func newWiredModel(metaType reflect.Type, target interface{}, proxy bool, priority int) *wiredModel {
	return &wiredModel{
		metaType: metaType,
		target:   target,
		proxy:    proxy,
		priority: priority,
	}
}
