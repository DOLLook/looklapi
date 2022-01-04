package irisserver_middleware

import (
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"micro-webapi/common/utils"
	"micro-webapi/errs"
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
	controllerType reflect.Type   // 控制器类型元素据
	paramValidator *reflect.Value // 参数验证
}

var ctrDataMap = make(map[string]*controllerMetadata)

// 注册控制器
// apiParty 路由组
// routePath 路由
// httpMethod 请求方式
// controller 控制器
// paramValidator 控制器参数校验
func RegisterController(irisApp *iris.Application, apiParty string, routePath string, httpMethod string, controller interface{}, paramValidator interface{}) {

	if irisApp == nil {
		panic(errors.New("irisapp must not be nil"))
	}

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

	if controller == nil {
		panic(errors.New("controller must not be nil"))
	}

	ctrVal := reflect.ValueOf(controller)
	if ctrVal.Kind() != reflect.Func {
		panic(errors.New("controller must be a func"))
	}

	ctrType := reflect.TypeOf(controller)
	for i := 0; i < ctrType.NumIn(); i++ {
		pk := ctrType.In(i).Kind()
		if pk == reflect.Slice || pk == reflect.Map || pk == reflect.Struct {
			continue
		}

		switch pk {
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64,
			reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
			break
		default:
			panic(errors.New("url param must be primitive type"))
		}
	}

	var pvalidator reflect.Value
	if paramValidator != nil {
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
	}

	party := irisApp.Party(apiParty)

	// 绑定路由
	switch httpMethod {
	case http.MethodGet:
		party.Get(routePath, apiSrvHandler)
		break
	case http.MethodHead:
		party.Head(routePath, apiSrvHandler)
		break
	case http.MethodPost:
		party.Post(routePath, apiSrvHandler)
		break
	case http.MethodOptions:
		party.Options(routePath, apiSrvHandler)
		break
	default:
		panic("not suppored the http method")
	}

	ctrMetadata := &controllerMetadata{
		party:          apiParty,
		routePath:      routePath,
		httpMethod:     httpMethod,
		controller:     &ctrVal,
		paramValidator: &pvalidator,
		controllerType: ctrType,
	}

	mkey := fmt.Sprintf("%s:%s%s", httpMethod, apiParty, routePath)
	ctrDataMap[mkey] = ctrMetadata
}

// 接口处理器
func apiSrvHandler(ctx iris.Context) {
	path := ctx.Path()
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "https://")

	startIndex := strings.Index(path, "/")
	endIndex := strings.Index(path, "?")

	var mkey string
	if endIndex >= 0 {
		mkey = fmt.Sprintf("%s:%s", ctx.Method(), path[startIndex:endIndex-1])
	} else {
		mkey = fmt.Sprintf("%s:%s", ctx.Method(), path[startIndex:])
	}

	ctrMetadata := ctrDataMap[mkey]
	if ctrDataMap == nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.Next()
		return
	}

	if ctx.Method() != ctrMetadata.httpMethod {
		err := fmt.Sprintf("%s not allowed %s", path, ctx.Method())
		ctx.SetErr(errs.NewBllError(err))
		ctx.Next()
		return
	}

	nonReqParamCount := 0
	ctrParams := make([]reflect.Value, 0)
	ctxUrlParams := ctx.URLParamsSorted()
	for i := 0; i < ctrMetadata.controllerType.NumIn(); i++ {
		paramType := ctrMetadata.controllerType.In(i)
		pk := paramType.Kind()
		if pk == reflect.Slice || pk == reflect.Map || pk == reflect.Struct {
			nonReqParamCount++

			val := reflect.New(paramType)
			if err := ctx.ReadJSON(val.Interface()); err != nil {
				ctx.SetErr(err)
				ctx.Next()
				return
			}

			ctrParams = append(ctrParams, val.Elem())
		} else {
			index := i - nonReqParamCount
			if len(ctxUrlParams) > index && !utils.IsEmpty(ctxUrlParams[index].Value) {
				if val, err := parseUrlParam(pk, ctxUrlParams[index].Value); err != nil {
					ctx.SetErr(err)
					ctx.Next()
					return
				} else {
					ctrParams = append(ctrParams, val)
				}
			} else {
				//ctrParams = append(ctrParams, reflect.ValueOf(nil))
				ctrParams = append(ctrParams, reflect.Zero(paramType))
			}
		}
	}

	if ctrMetadata.paramValidator != nil {
		validateResult := ctrMetadata.paramValidator.Call(ctrParams)
		if validateResult != nil && len(validateResult) > 0 {
			if err, ok := validateResult[0].Interface().(error); ok {
				ctx.SetErr(err)
				ctx.Next()
				return
			}
		}
	}

	ctrMetadata.controller.Call(ctrParams)
	ctx.Next()
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
