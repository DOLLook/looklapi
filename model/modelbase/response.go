package modelbase

import "micro-webapi/errs"

// api错误响应
type errResponse struct {
	// 是否成功
	IsSuccess bool
	// 错误码
	ErrorCode int
	// 错误信息
	ErrorMsg string
}

// api请求响应值
type responseResult struct {
	errResponse
	// 结果
	Result interface{}
}

// 请求结果
func NewResponse(data interface{}) (result *responseResult) {
	result = &responseResult{
		Result: data,
	}
	result.IsSuccess = true
	return
}

// 错误响应
func NewErrResponse(err interface{}) *errResponse {
	eresp := &errResponse{
		IsSuccess: false,
		ErrorCode: -1,
	}

	if err == nil {
		return eresp
	}

	if err, ok := err.(error); ok {
		if berr, ok := err.(*errs.BllError); ok {
			eresp.ErrorCode = berr.Code
			eresp.ErrorMsg = berr.Msg
		} else {
			eresp.ErrorMsg = err.Error()
		}
	}

	return eresp
}
