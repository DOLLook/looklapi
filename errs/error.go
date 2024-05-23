package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// 业务错误类型
type BllError struct {
	Code       int    `json:"ErrorCode"`
	Msg        string `json:"ErrorMsg"`
	Tip        string
	callers    []frame        // 调用堆栈
	stackTrace []*formatStack // 格式化堆栈
}

func (e *BllError) Error() string {
	tempTip := strings.TrimSpace(e.Tip)
	if len(tempTip) == 0 {
		return fmt.Sprintf("Code: %d, Msg: %s", e.Code, e.Msg)
	} else {
		return fmt.Sprintf("Code: %d, Msg: %s, Tip: %s", e.Code, e.Msg, e.Tip)
	}
}

// 获取堆栈
func (e *BllError) FormatStackTrace() []*formatStack {
	return e.stackTrace
}

// 获取堆栈
func (e *BllError) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(e.callers))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((e.callers)[i])
	}
	return f
}

func NewBllError(msg string) error {
	//return errors.WithStack(&BllError{Code: -1, Msg: msg})
	err := &BllError{Code: -1, Msg: msg, callers: callers()}
	err.generateStackTrace()
	return err
}

func NewBllErrorWithCode(msg string, code int) error {
	//return errors.WithStack(&BllError{Code: code, Msg: msg})
	err := &BllError{Code: code, Msg: msg, callers: callers()}
	err.generateStackTrace()
	return err
}

func NewBllErrorWithCodeTip(msg string, code int, tip string) error {
	//return errors.WithStack(&BllError{Code: code, Msg: msg, Tip: tip})
	err := &BllError{Code: code, Msg: msg, Tip: tip, callers: callers()}
	err.generateStackTrace()
	return err
}

func (e *BllError) generateStackTrace() {
	e.stackTrace = make([]*formatStack, len(e.callers))
	for index, item := range e.callers {
		e.stackTrace[index] = item.toStack()
	}
}
