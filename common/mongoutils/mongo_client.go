package mongoutils

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonoptions"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"looklapi/common/utils"
	appConfig "looklapi/config"
	"reflect"
	"time"
)

const (
	_DBNAME = "test"
)

var mongoUri = appConfig.AppConfig.MongodbUri
var mongoClient *mongo.Client
var clientInitialized = false

func init() {
	if utils.IsEmpty(mongoUri) {
		return
	}
	timeDecoder := bsoncodec.NewTimeCodec(bsonoptions.TimeCodec().SetUseLocalTimeZone(true))
	registry := bson.NewRegistryBuilder().
		RegisterCodec(reflect.TypeOf((*decimal.Decimal)(nil)).Elem(), mongoDecimal{}).
		RegisterTypeDecoder(reflect.TypeOf((*time.Time)(nil)).Elem(), timeDecoder).
		Build()

	client, _ := mongo.Connect(context.TODO(),
		options.Client().ApplyURI(mongoUri).SetRegistry(registry))

	mongoClient = client
	clientInitialized = true
}

func ClientIsValid() bool {
	return clientInitialized
}

// 获取数据库
func GetDatabase(dbName string) *mongo.Database {
	if utils.IsEmpty(dbName) {
		dbName = _DBNAME
	}
	return mongoClient.Database(dbName)
}

// 获取集合
func GetCollection(collName string) *mongo.Collection {
	return GetDatabase(_DBNAME).Collection(collName)
}

// 获取client
func GetMongoClient() *mongo.Client {
	return GetDatabase(_DBNAME).Client()
}

type mongoDecimal decimal.Decimal

func (d mongoDecimal) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	decimalType := reflect.TypeOf((*decimal.Decimal)(nil)).Elem()
	if !val.IsValid() || !val.CanSet() || val.Type() != decimalType {
		return bsoncodec.ValueDecoderError{
			Name:     "decimalDecodeValue",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	var value decimal.Decimal
	switch vr.Type() {
	case bsontype.Decimal128:
		dec, err := vr.ReadDecimal128()
		if err != nil {
			return err
		}
		value, err = decimal.NewFromString(dec.String())
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("received invalid BSON type to decode into decimal.Decimal: %s", vr.Type())
	}

	val.Set(reflect.ValueOf(value))
	return nil
}

func (d mongoDecimal) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	decimalType := reflect.TypeOf((*decimal.Decimal)(nil)).Elem()
	if !val.IsValid() || val.Type() != decimalType {
		return bsoncodec.ValueEncoderError{
			Name:     "decimalEncodeValue",
			Types:    []reflect.Type{decimalType},
			Received: val,
		}
	}

	dec := val.Interface().(decimal.Decimal)
	dec128, err := primitive.ParseDecimal128(dec.String())
	if err != nil {
		return err
	}

	return vw.WriteDecimal128(dec128)
}
