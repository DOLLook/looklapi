package errs

import (
	"fmt"
)

// 业务错误类型
type BllError struct {
	Code int    `json:"ErrorCode"`
	Msg  string `json:"ErrorMsg"`
	Tip  string
}

func (e *BllError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s, tip: %s", e.Code, e.Msg, e.Tip)
}

func New(code int, msg string, tip string) *BllError {
	return &BllError{code, msg, tip}
}
