package model

// 表集合接口
type DbTable interface {
	// 表或集合名称
	TbCollName() string
}
