package mongoutils

import (
	"context"
	"go-webapi-fw/common/utils"
	appConfig "go-webapi-fw/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DBNAME = "hi-nature"
)

var mongoUri = appConfig.AppConfig.MongodbUri
var mongoClient *mongo.Client

func init() {
	client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	mongoClient = client
}

// 获取数据库
func GetDatabase(dbName string) *mongo.Database {
	if utils.IsEmpty(dbName) {
		dbName = DBNAME
	}
	return mongoClient.Database(dbName)
}

// 获取集合
func GetCollection(collName string) *mongo.Collection {
	return GetDatabase(DBNAME).Collection(collName)
}
