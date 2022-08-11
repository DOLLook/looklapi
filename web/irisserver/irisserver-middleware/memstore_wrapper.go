package irisserver_middleware

import (
	"github.com/kataras/iris/v12/core/memstore"
)

type memStoreWrapper struct {
	*memstore.Store
}

func (w *memStoreWrapper) Exists(key string) bool {
	return w.Store.Exists(key)
}

func (w *memStoreWrapper) Get(key string) interface{} {
	return w.Store.Get(key)
}

func (w *memStoreWrapper) Save(key string, value interface{}, immutable bool) bool {
	_, result := w.Store.Save(key, value, immutable)
	return result
}

func (w *memStoreWrapper) Remove(key string) bool {
	return w.Store.Remove(key)
}
