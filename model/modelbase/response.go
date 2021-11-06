package modelbase

// api请求响应值
type responseResult struct {
	// 是否成功
	IsSuccess bool
	// 错误码
	ErrorCode int
	// 错误信息
	ErrorMsg string
	// 结果
	Result interface{}
}

// 请求结果
func NewResponse(data interface{}) (result *responseResult) {
	result = &responseResult{
		IsSuccess: true,
		Result:    data,
	}
	return
}
