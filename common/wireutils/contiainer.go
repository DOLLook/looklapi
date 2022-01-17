package wireutils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

var container = make(map[reflect.Type][]*wiredModel)
var mu = &sync.Mutex{}
var injected = false

// 映射接口实例
func Bind(itype reflect.Type, target interface{}, proxy bool, priority int) {
	if target == nil {
		panic("target must not be nil")
	}

	ttype := reflect.ValueOf(target)
	tk := ttype.Kind()
	if tk != reflect.Ptr {
		panic("target must be a struct pointer")
	}

	tk = ttype.Elem().Kind()
	if tk != reflect.Struct {
		panic("target must be a struct pointer")
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

	proxyIndex := -1
	for i, t := range targets {
		if t.proxy {
			proxyIndex = i
			break
		}
	}
	//proxyIndex := linq.From(targets).IndexOf(func(item interface{}) bool {
	//	return item.(*wiredModel).proxy
	//})

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
	targets := resovle(itype)
	return targets[0].target
}

// 注入对象
func Inject() {
	if injected {
		return
	}

	for _, slice := range container {
		for _, wiredModel := range slice {
			inject(wiredModel)
		}
	}

	injected = true
}

func inject(model *wiredModel) {
	if model.injected {
		return
	}

	tval := reflect.ValueOf(model.target).Elem()
	ttype := reflect.TypeOf(model.target).Elem()
	nfield := ttype.NumField()
	if nfield <= 0 {
		model.injected = true
		return
	}

	model.injecting = true
	for i := 0; i < nfield; i++ {
		field := ttype.Field(i)
		val, ok := field.Tag.Lookup("wired")
		if !ok {
			continue
		}

		if strings.ToLower(strings.TrimSpace(val)) != "autowired" {
			continue
		}

		index := -1
		ftyp := field.Type
		children := resovle(ftyp)
		for ich, child := range children {
			if child.injecting {
				continue
			}

			inject(child)

			if model.metaType == ftyp {
				if !child.proxy {
					// 不支持多层代理
					index = ich
					break
				}
			} else {
				index = ich
				break
			}
		}

		if index < 0 {
			panic(fmt.Sprintf("can not resolve the type %s", ftyp.Name()))
		}

		fieldVal := tval.Field(i)
		ptr := reflect.NewAt(ftyp, unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		ptr.Set(reflect.ValueOf(children[index].target))
	}

	model.injecting = false
	model.injected = true
}

// 获取对象
func resovle(itype reflect.Type) []*wiredModel {
	if len(container) <= 0 {
		panic(fmt.Sprintf("can not resolve the type %s", itype.Name()))
	}

	targets := container[itype]
	if len(targets) <= 0 {
		panic(fmt.Sprintf("can not resolve the type %s", itype.Name()))
	}

	return targets
}
