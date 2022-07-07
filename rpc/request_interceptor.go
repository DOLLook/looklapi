package rpc

import (
	"looklapi/common/utils"
	"net/http"
)

// 请求模板数据
type requestTemplate struct {
	header        *http.Header // 请求头
	body          interface{}  // post formdata
	urlParamSlice []string     // url参数 slice item: key=vale
}

// 添加请求头
func (t *requestTemplate) appendHeader(name string, val string) {
	if utils.IsEmpty(name) || utils.IsEmpty(val) || t.header == nil {
		return
	}

	t.header.Add(name, val)
}

// 设置请求头
func (t *requestTemplate) setHeader(name string, val string) {
	if utils.IsEmpty(name) || utils.IsEmpty(val) || t.header == nil {
		return
	}

	t.header.Set(name, val)
}

// 获取请求头
func (t *requestTemplate) getHeader(name string) string {
	if utils.IsEmpty(name) || t.header == nil {
		return ""
	}

	return t.header.Get(name)
}

type interceptor func(*requestTemplate)

var gloabalReqInterceptor = []interceptor{

	// 示例: 在此构造全局统一请求头
	func(template *requestTemplate) {
		if template == nil {
			return
		}

		if template.header == nil {
			template.header = &http.Header{}
		}

		//agent := "looklapi"
		//link := template.getHeader("yourHeaderName")
		//if !utils.IsEmpty(link) {
		//	agent = link + "," + agent
		//}
		//
		//template.setHeader("yourHeaderName", agent)
	},
}
