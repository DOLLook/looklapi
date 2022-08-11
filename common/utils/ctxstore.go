package utils

// 上下文存储接口
type CtxStore interface {
	Exists(key string) bool
	Get(key string) interface{}
	Save(key string, value interface{}, immutable bool) bool
	Remove(key string) bool
}
