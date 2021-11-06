package mongoutils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// 系统日志
type systemRuntineLog struct {
	Id string `bson:"_id"`

	// 实例名
	Instance string `bson:"instance"`

	// 线程id
	ThreadId int32 `bson:"thread_id"`

	// 运行类名
	ClassName string `bson:"class_name"`

	// 日志等级
	Level int32 `bson:"level"`

	// 宿主IP
	HostIp string `bson:"host_ip"`

	// 时间
	Time time.Time `bson:"time"`

	// 日志内容
	Content string `bson:"content"`

	// 堆栈信息
	Stacktrace string `bson:"stacktrace"`
}

// 获取集合名称
func (log *systemRuntineLog) TbCollName() string {
	return "sys_runtine_log"
}

// 新建日志
func NewMongoLog() *systemRuntineLog {
	return &systemRuntineLog{
		Id:   primitive.NewObjectID().Hex(),
		Time: time.Now(),
	}
}
