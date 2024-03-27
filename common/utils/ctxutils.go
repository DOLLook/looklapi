package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

// 上下文存储接口
type CtxStore interface {
	Exists(key string) bool
	Get(key string) interface{}
	Save(key string, value interface{}, immutable bool) bool
	Remove(key string) bool
}

// 通过http context获取请求头对象
func GetHttpHeader(httpCtx context.Context) http.Header {
	if httpCtx == nil {
		return nil
	}
	header := httpCtx.Value(HttpRequestHeader)
	if header == nil {
		return nil
	}

	httpHeader, ok := header.(http.Header)
	if !ok {
		return nil
	}

	return httpHeader
}

// 通过http context获取请求头
func GetHttpHeaderVal(httpCtx context.Context, header string) string {
	headerMap := GetHttpHeader(httpCtx)
	if headerMap == nil {
		return ""
	}

	return headerMap.Get(header)
}

// 通过http context获取上下文临时存储
func GetHttpCtxStore(httpCtx context.Context) CtxStore {
	if httpCtx == nil {
		return nil
	}

	store := httpCtx.Value(HttpContextStore)
	if store == nil {
		return nil
	}

	if store, ok := store.(CtxStore); ok {
		return store
	} else {
		return nil
	}
}

var NotFoundCtxValueErr = errors.New("not found context value")

// 获取context value
func GetCtxValue[T any](ctx context.Context, key any) (T, error) {
	var result T
	if key == nil {
		return result, errors.New("key must not nil")
	}

	if ctxVal := ctx.Value(key); ctxVal != nil {
		if v, ok := ctxVal.(T); ok {
			return v, nil
		} else {
			return result, errors.New(fmt.Sprintf("context value with type %s not match the type %s", reflect.TypeOf(ctxVal).String(), reflect.TypeOf(result).String()))
		}
	} else {
		return result, NotFoundCtxValueErr
	}
}
