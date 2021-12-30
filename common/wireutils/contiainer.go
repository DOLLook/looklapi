package wireutils

import (
	"errors"
	"github.com/ahmetb/go-linq/v3"
	"reflect"
	"sort"
	"sync"
)

var container = make(map[reflect.Type][]*wiredModel)
var mu = &sync.Mutex{}

// 映射接口实例
func Bind(itype reflect.Type, target interface{}, proxy bool, priority int) {
	if target == nil {
		panic(errors.New("target must not be nil"))
	}

	ttype := reflect.ValueOf(target)
	tk := ttype.Kind()
	if tk != reflect.Ptr && tk != reflect.Uintptr && tk != reflect.UnsafePointer {
		panic(errors.New("target must be a pointer"))
	}

	mu.Lock()
	defer mu.Unlock()

	wm := newWiredModel(itype, target, proxy, priority)
	container[itype] = append(container[itype], wm)

	if len(container[itype]) == 1 {
		return
	}

	targets := container[itype]
	temp := make([]*wiredModel, 0)

	proxyIndex := linq.From(targets).IndexOf(func(item interface{}) bool {
		return item.(*wiredModel).proxy
	})
	if proxyIndex >= 0 {
		temp = append(temp, targets[proxyIndex])
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].priority <= targets[i].priority
	})

	for _, tg := range targets {
		if tg.proxy {
			continue
		}

		temp = append(temp, tg)
	}

	container[itype] = temp
}

// 获取对象
func Resovle(itype reflect.Type) interface{} {
	if len(container) <= 0 {
		panic(errors.New("can not resolve this type"))
	}

	targets := container[itype]
	if len(targets) <= 0 {
		panic(errors.New("can not resolve this type"))
	}

	return targets[0].target
}
