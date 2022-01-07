package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"micro-webapi/common/loggers"
	service_discovery "micro-webapi/common/service-discovery"
	"micro-webapi/common/utils"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type RpcService interface {
	SrvName() string
}

var srvMap = make(map[string]RpcService)

var headerType = reflect.TypeOf((*http.Header)(nil)).Elem()

// 获取调用客户端
func GetHttpRpcClient(srvName utils.EnumServiceName) RpcService {
	return srvMap[string(srvName)]
}

// register service
func register(srv RpcService) {
	if srv == nil || utils.IsEmpty(srv.SrvName()) {
		return
	}

	if srvMap[srv.SrvName()] != nil {
		return
	}

	srvGenerator(srv)
	srvMap[srv.SrvName()] = srv
}

func srvGenerator(srv RpcService) {

	srvVal := reflect.ValueOf(srv)
	if srvVal.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("service: %s, must be registered as a ptr", srv.SrvName()))
	}

	srvVal = srvVal.Elem()
	srvTyp := srvVal.Type()

	for i := 0; i < srvTyp.NumField(); i++ {
		fn := srvTyp.Field(i)
		reqMethod, reqRoute, aliasSlice := apiDefCheck(fn, srv.SrvName())
		fnWrap := reflect.MakeFunc(fn.Type, httpRpcWrap(srv.SrvName(), fn.Type, reqMethod, reqRoute, aliasSlice))
		srvVal.Field(i).Set(fnWrap)
	}
}

func httpRpcWrap(srvName string, fntyp reflect.Type, reqMethod string, reqRoute string, aliasSlice []string) func(args []reflect.Value) (results []reflect.Value) {

	return func(args []reflect.Value) (results []reflect.Value) {

		defer func() {
			if err := recover(); err != nil {
				var errResult error
				if tr, ok := err.(error); ok {
					errResult = tr
					loggers.GetLogger().Error(errResult)
				} else if msg, ok := err.(string); ok {
					errResult = errors.New(msg)
					loggers.GetLogger().Error(errResult)
				}

				results = append(results, reflect.ValueOf(errResult))
			}
		}()

		arglen := len(args)
		respReciever := args[arglen-1]
		if respReciever.IsNil() {
			results = append(results, reflect.ValueOf(errors.New("response result reciever can not be null pointer")))
			return
		}

		serviceName := srvName
		funtyp := fntyp
		method := reqMethod
		route := reqRoute
		alias := aliasSlice
		header, body, urlParamSlice := reqParamGenerator(funtyp, args, alias)

		url, err := getEndpoint(serviceName)
		if err != nil {
			results = append(results, reflect.ValueOf(err))
			return
		}
		url = url + route

		if resp, err := doRequest(method, url, header, body, urlParamSlice); err != nil {
			results = append(results, reflect.ValueOf(err))
			return
		} else if err := json.Unmarshal(resp, respReciever.Interface()); err != nil {
			results = append(results, reflect.ValueOf(err))
			return
		}

		results = append(results, reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()))
		return
	}
}

// check the definition of the api
func apiDefCheck(fn reflect.StructField, srvName string) (method, route string, alias []string) {
	if fn.Type.Kind() != reflect.Func {
		panic(fmt.Sprintf("service: %s, %s must be a fun", srvName, fn.Name))
	}

	fnTag := fn.Tag
	reqMethod, reqRoute := tagCheck(fnTag, srvName, fn.Name)

	outCheck(fn, srvName)

	urlParamCount, hasBody := inCheck(fn, srvName)
	if hasBody && reqMethod != http.MethodPost {
		panic(fmt.Sprintf("service: %s, func: %s, the request that has a body param must use POST", srvName, fn.Name))
	}

	aliasSlice := getParamAlias(fnTag)
	if urlParamCount != len(aliasSlice) {
		panic(fmt.Sprintf("service: %s, func: %s, the request param that exclued the header and body must match param alias count",
			srvName, fn.Name))
	}

	method, route, alias = reqMethod, reqRoute, aliasSlice
	return
}

func tagCheck(fnTag reflect.StructTag, srvName string, fnName string) (reqMethod, reqRoute string) {

	reqMethod, ok := fnTag.Lookup("method")
	if !ok {
		panic(fmt.Sprintf("service: %s, func: %s, not specify the request method", srvName, fnName))
	} else {
		reqMethod = strings.ToUpper(reqMethod)
	}

	reqRoute, ok = fnTag.Lookup("route")
	if !ok {
		panic(fmt.Sprintf("service: %s, func: %s, not specify the request route", srvName, fnName))
	}

	if reqMethod != http.MethodGet && reqMethod != http.MethodPost {
		panic(fmt.Sprintf("service: %s, func: %s, request method must be GET or POST", srvName, fnName))
	}

	return
}

func outCheck(fn reflect.StructField, srvName string) {
	outCount := fn.Type.NumOut()
	if outCount < 1 {
		panic(fmt.Sprintf("service: %s, func: %s, must set an error return", srvName, fn.Name))
	}

	if outCount > 1 {
		panic(fmt.Sprintf("service: %s, func: %s, must set only one error return", srvName, fn.Name))
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if fn.Type.Out(0) != errorType {
		panic(fmt.Sprintf("service: %s, func: %s, return must be error", srvName, fn.Name))
	}
}

func inCheck(fn reflect.StructField, srvName string) (urlParamCount int, hasBody bool) {
	incount := fn.Type.NumIn()
	if incount < 1 {
		panic(fmt.Sprintf("service: %s, func: %s, must set an interface ptr to recieve the response result", srvName, fn.Name))
	}

	respRecieverTyp := fn.Type.In(incount - 1)
	if respRecieverTyp.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("service: %s, func: %s, the response result reciever must be an interface ptr", srvName, fn.Name))
	}

	respRecieverTyp = respRecieverTyp.Elem()
	respRecieverkind := respRecieverTyp.Kind()
	if respRecieverkind != reflect.Struct && respRecieverkind != reflect.Slice && respRecieverkind != reflect.Map {
		panic(fmt.Sprintf("service: %s, func: %s, the response result reciever must be an interface ptr", srvName, fn.Name))
	}

	urlParamCount = 0
	hasBody = false
	for i := 0; i < incount-1; i++ {
		isPtr := false
		ptype := fn.Type.In(i)
		if ptype.Kind() == reflect.Ptr {
			isPtr = true
			ptype = ptype.Elem()
		}

		if ptype == headerType {
			continue
		}

		ptk := ptype.Kind()
		if ptk == reflect.Struct || ptk == reflect.Slice || ptk == reflect.Map {
			if hasBody {
				panic(fmt.Sprintf("service: %s, func: %s, request body can not set more than one", srvName, fn.Name))
			}
			hasBody = true
			continue
		} else if isPtr {
			panic(fmt.Sprintf("service: %s, func: %s, url param must be primitive type", srvName, fn.Name))
		}

		switch ptk {
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64,
			reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
			urlParamCount++
		default:
			panic(fmt.Sprintf("service: %s, func: %s, url param must be primitive type", srvName, fn.Name))
		}
	}

	return
}

func getParamAlias(fnTag reflect.StructTag) []string {
	paramAlias, aliasOk := fnTag.Lookup("alias")
	aliasSlice := make([]string, 0)
	if aliasOk {
		paramAlias = strings.TrimPrefix(paramAlias, "[")
		paramAlias = strings.TrimSuffix(paramAlias, "]")
		for _, alias := range strings.Split(paramAlias, ",") {
			aliasSlice = append(aliasSlice, strings.TrimSpace(alias))
		}
	}

	return aliasSlice
}

// generate the request params
func reqParamGenerator(fntyp reflect.Type, args []reflect.Value, alias []string) (header *http.Header, body interface{}, urlParamSlice []string) {

	for i := 0; i < len(args)-1; i++ {
		ptyp := fntyp.In(i)
		arg := args[i]

		if ptyp.Kind() == reflect.Ptr {
			ptyp = ptyp.Elem()
			arg = arg.Elem()
		}

		if ptyp == headerType {
			if !arg.IsNil() {
				h := arg.Interface().(http.Header)
				header = &h
			}
			continue
		}

		ptk := ptyp.Kind()
		if ptk == reflect.Struct || ptk == reflect.Slice || ptk == reflect.Map {
			body = arg.Interface()
			continue
		}

		if !arg.IsZero() {
			urlParamSlice = append(urlParamSlice, fmt.Sprintf("%s=%v", alias[i], arg.Interface()))
		}
	}

	return
}

func getEndpoint(srvName string) (endpoint string, err error) {
	srvslice := service_discovery.GetServiceManager().GetHealthServices(srvName)
	srvlen := len(srvslice)
	if srvlen < 1 {
		return "", errors.New("can't find health service")
	} else if srvlen == 1 {
		return "http://" + srvslice[0], nil
	} else {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		idx := r.Intn(srvlen)
		return "http://" + srvslice[idx], nil
	}
}
