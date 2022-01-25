package irisserver_middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"looklapi/common/utils"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// 控制器元数据
type controllerMetadata struct {
	party          string         // 路由组
	routePath      string         // 路由
	httpMethod     string         // 方法
	controller     *reflect.Value // 控制器
	controllerType reflect.Type   // 控制器类型元数据
	paramValidator *reflect.Value // 参数验证
}

var ctrDataMap = make(map[string]*controllerMetadata)

// 注册控制器
// apiParty 路由组
// routePath 路由
// httpMethod 请求方式
// controller 控制器
// paramValidator 控制器参数校验
func RegisterController(irisApp *iris.Application, apiParty string, routePath string, httpMethod string, controller interface{}, paramValidator interface{}, beforeHandlers []iris.Handler, afterHandlers []iris.Handler) {

	if irisApp == nil {
		panic(errors.New("irisapp must not be nil"))
	}

	apiParty, routePath = routeCheck(apiParty, routePath)
	mkey := fmt.Sprintf("%s:%s%s", httpMethod, apiParty, routePath)
	if ctrDataMap[mkey] != nil {
		panic(errors.New(mkey + " duplicated register"))
	}

	ctrValuePtr, ctrType := controllerCheck(controller)
	paramValidatorValuePtr := paramValidatorCheck(paramValidator, ctrType)

	_ = registerRoute(irisApp, apiParty, routePath, httpMethod, beforeHandlers, afterHandlers)

	ctrMetadata := &controllerMetadata{
		party:          apiParty,
		routePath:      routePath,
		httpMethod:     httpMethod,
		controller:     ctrValuePtr,
		paramValidator: paramValidatorValuePtr,
		controllerType: ctrType,
	}

	ctrDataMap[mkey] = ctrMetadata
}

// 接口处理器
func apiSrvHandler(ctx iris.Context) {
	relativePath, mkey := reqApiMapKey(ctx)

	ctrMetadata := ctrDataMap[mkey]
	if ctrDataMap == nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.Next()
		return
	}

	if ctx.Method() != ctrMetadata.httpMethod {
		err := fmt.Sprintf("%s not allowed %s", relativePath, ctx.Method())
		ctx.SetErr(errors.New(err))
		ctx.Next()
		return
	}

	ctrParams, err := reqApiParams(ctx, ctrMetadata)
	if err != nil {
		ctx.SetErr(err)
		ctx.Next()
		return
	}

	if err := validate(ctrMetadata, ctrParams); err != nil {
		ctx.SetErr(err)
		ctx.Next()
		return
	}

	if resp, err := handle(ctrMetadata, ctrParams); err != nil {
		ctx.SetErr(err)
		ctx.Next()
		return
	} else if resp != nil {
		if _, err := ctx.JSON(resp); err != nil {
			ctx.SetErr(err)
			ctx.Next()
			return
		}
	}

	ctx.Next()
}

func routeCheck(apiParty string, routePath string) (string, string) {
	if utils.IsEmpty(apiParty) {
		panic(errors.New("apiParty must not be nil or empty"))
	}

	if !strings.HasPrefix(apiParty, "/") {
		apiParty = "/" + apiParty
	}

	if utils.IsEmpty(routePath) {
		panic(errors.New("routePath must not be nil or empty"))
	}

	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}

	return apiParty, routePath
}

func controllerCheck(controller interface{}) (*reflect.Value, reflect.Type) {
	if controller == nil {
		panic(errors.New("controller must not be nil"))
	}

	ctrVal := reflect.ValueOf(controller)
	if ctrVal.Kind() != reflect.Func {
		panic(errors.New("controller must be a func"))
	}

	hasBody := false
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	ctrType := reflect.TypeOf(controller)
	for i := 0; i < ctrType.NumIn(); i++ {
		isPtr := false
		paramType := ctrType.In(i)
		if paramType.Kind() == reflect.Ptr {
			isPtr = true
			paramType = paramType.Elem()
		}

		if paramType == contextType {
			if i != 0 {
				panic(errors.New("context.Context must be the first param"))
			} else if isPtr {
				panic(errors.New("context.Context can not be ptr"))
			} else {
				continue
			}
		}

		pk := paramType.Kind()
		if pk == reflect.Slice || pk == reflect.Map || pk == reflect.Struct {
			if hasBody {
				panic(errors.New("request body can not set more than one"))
			}
			hasBody = true
			continue
		}

		if isPtr {
			panic(errors.New("url param must be primitive type"))
		}

		switch pk {
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64,
			reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
			break
		default:
			panic(errors.New("url param must be primitive type"))
		}
	}

	outCount := ctrType.NumOut()
	if outCount == 1 {
		// only return error
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if ctrType.Out(0) != errorType {
			panic(errors.New("controller's only one return must be error interface"))
		}
	} else if outCount == 2 {
		// return result and error, the first one is result, the second one is error
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if ctrType.Out(1) != errorType {
			panic(errors.New("controller's last return must be error interface"))
		}
	} else {
		panic(errors.New("controller's return must be one or two. less than 1 or more than 2 return is not supported"))
	}

	return &ctrVal, ctrType
}

func paramValidatorCheck(paramValidator interface{}, ctrType reflect.Type) *reflect.Value {
	if paramValidator == nil {
		return nil
	}

	var pvalidator reflect.Value
	pvalidator = reflect.ValueOf(paramValidator)
	if pvalidator.Kind() != reflect.Func {
		panic(errors.New("paramValidator must be a func"))
	}

	pvalidatorType := reflect.TypeOf(paramValidator)

	if ctrType.NumIn() != pvalidatorType.NumIn() {
		panic(errors.New("paramValidator's param must be the same of controller"))
	}

	if pvalidatorType.NumOut() != 1 {
		panic(errors.New("paramValidator's return must be one"))
	}

	//outKind := pvalidatorType.Out(0).Kind()
	//if outKind != reflect.Interface {
	//	panic(errors.New("paramValidator's return must be error"))
	//}
	//mustImplType := reflect.TypeOf((*error)(nil)).Elem()
	//if !pvalidatorType.Out(0).Implements(mustImplType) {
	//	panic(errors.New("paramValidator's return must implements error interface"))
	//}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if pvalidatorType.Out(0) != errorType {
		panic(errors.New("paramValidator's return must be error interface"))
	}

	for i := 0; i < ctrType.NumIn(); i++ {
		ctrParam := ctrType.In(i)
		validatorParam := pvalidatorType.In(i)

		if ctrParam != validatorParam {
			panic(errors.New("paramValidator's param must be the same of controller"))
		}
	}

	return &pvalidator
}

func registerRoute(irisApp *iris.Application, apiParty string, routePath string, httpMethod string, beforeHandlers []iris.Handler, afterHandlers []iris.Handler) iris.Party {
	party := irisApp.Party(apiParty)

	var handlerSlice []iris.Handler
	if len(beforeHandlers) > 0 {
		handlerSlice = append(handlerSlice, beforeHandlers...)
	}
	handlerSlice = append(handlerSlice, apiSrvHandler)
	if len(afterHandlers) > 0 {
		handlerSlice = append(handlerSlice, afterHandlers...)
	}

	// 绑定路由
	switch httpMethod {
	case http.MethodGet:
		party.Get(routePath, handlerSlice...)
		break
	case http.MethodHead:
		party.Head(routePath, handlerSlice...)
		break
	case http.MethodPost:
		party.Post(routePath, handlerSlice...)
		break
	case http.MethodOptions:
		party.Options(routePath, handlerSlice...)
		break
	default:
		panic("not suppored the http method")
	}

	return party
}

// get the controller metadata key
func reqApiMapKey(ctx iris.Context) (relativePath, mkey string) {
	path := ctx.Path()
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "https://")

	startIndex := strings.Index(path, "/")
	endIndex := strings.Index(path, "?")

	if endIndex >= 0 {
		relativePath = path[startIndex : endIndex-1]
	} else {
		relativePath = path[startIndex:]
	}

	mkey = fmt.Sprintf("%s:%s", ctx.Method(), relativePath)

	return
}

// get the request param's reflet.Value slice
func reqApiParams(ctx iris.Context, ctrMetadata *controllerMetadata) ([]reflect.Value, error) {
	nonReqParamCount := 0
	ctrParams := make([]reflect.Value, 0)
	ctxUrlParams := ctx.URLParamsSorted()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	for i := 0; i < ctrMetadata.controllerType.NumIn(); i++ {
		isPtr := false
		paramType := ctrMetadata.controllerType.In(i)
		if paramType.Kind() == reflect.Ptr {
			isPtr = true
			paramType = paramType.Elem()
		}

		if paramType == contextType {
			nonReqParamCount++

			myCtx := context.WithValue(context.Background(), utils.HttpRequestHeader, ctx.Request().Header)
			for _, entry := range *ctx.Values() {
				myCtx = context.WithValue(myCtx, entry.Key, entry.Value())
			}

			ctrParams = append(ctrParams, reflect.ValueOf(myCtx))
			continue
		}

		pk := paramType.Kind()
		if pk == reflect.Slice || pk == reflect.Map || pk == reflect.Struct {
			nonReqParamCount++

			val := reflect.New(paramType)
			if err := ctx.ReadJSON(val.Interface()); err != nil {
				return nil, err
			}

			if isPtr {
				ctrParams = append(ctrParams, val)
			} else {
				ctrParams = append(ctrParams, val.Elem())
			}

		} else {
			index := i - nonReqParamCount
			if len(ctxUrlParams) > index && !utils.IsEmpty(ctxUrlParams[index].Value) {
				if val, err := parseUrlParam(pk, ctxUrlParams[index].Value); err != nil {
					return nil, err
				} else {
					ctrParams = append(ctrParams, val)
				}
			} else {
				//ctrParams = append(ctrParams, reflect.ValueOf(nil))
				ctrParams = append(ctrParams, reflect.Zero(paramType))
			}
		}
	}

	return ctrParams, nil
}

func parseUrlParam(paramKind reflect.Kind, strVal string) (reflect.Value, error) {
	switch paramKind {
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if val, err := strconv.Atoi(strVal); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(val), nil
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(strVal); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(val), nil
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(strVal, 32); err != nil {
			return reflect.ValueOf(nil), err
		} else {
			return reflect.ValueOf(val), nil
		}
	case reflect.String:
		return reflect.ValueOf(strVal), nil
	}

	return reflect.ValueOf(nil), nil
}

// do validate the params
func validate(ctrMetadata *controllerMetadata, ctrParams []reflect.Value) error {
	if ctrMetadata.paramValidator == nil {
		return nil
	}

	validateResult := ctrMetadata.paramValidator.Call(ctrParams)
	if validateResult != nil && len(validateResult) > 0 {
		if err := validateResult[0].Interface(); err != nil {
			return err.(error)
		}
	}
	return nil
}

// do process
func handle(ctrMetadata *controllerMetadata, ctrParams []reflect.Value) (interface{}, error) {
	handleResult := ctrMetadata.controller.Call(ctrParams)
	if handleResult != nil && len(handleResult) > 0 {
		if ctrMetadata.controllerType.NumOut() == 1 {
			if err := handleResult[0].Interface(); err != nil {
				return nil, err.(error)
			}
		} else {
			if err := handleResult[1].Interface(); err != nil {
				return nil, err.(error)
			}

			if resp := handleResult[0].Interface(); resp != nil {
				return resp, nil
			}
		}
	}

	return nil, nil
}
