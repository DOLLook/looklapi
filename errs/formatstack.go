package errs

import (
	"runtime"
	"strings"
)

// 格式化堆栈
type formatStack struct {
	file     string // 文件
	fileName string // 文件名
	method   string // func名
	line     int    // 行号
	invalid  bool   // 无效
}

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type frame uintptr

func callers() []frame {
	const depth = 32 // 默认最大深度32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	f := make([]frame, n)
	for i := 0; i < n; i++ {
		f[i] = frame(pcs[i])
	}

	return f
}

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f frame) pc() uintptr { return uintptr(f) - 1 }

// MarshalText formats a stacktrace Frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f frame) toStack() *formatStack {
	stack := &formatStack{}
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		stack.invalid = true
		stack.method = "unknown"
		return stack
	}

	stack.method = fn.Name()
	file, line := fn.FileLine(f.pc())

	indexCutNum := strings.Index(file, "/src/")
	trimLeftFile := file[indexCutNum+4:]
	temp := strings.Split(trimLeftFile, "/")
	fileName := temp[len(temp)-1]

	stack.file = trimLeftFile
	stack.fileName = fileName
	stack.line = line

	return stack
}

// 获取文件路径
func (s *formatStack) File() string {
	return s.file
}

// 获取文件名
func (s *formatStack) FileName() string {
	return s.fileName
}

// 获取方法
func (s *formatStack) Method() string {
	return s.method
}

// 获取行号
func (s *formatStack) Line() int {
	return s.line
}

// 取值是否有效
func (s *formatStack) Invalid() bool {
	return s.invalid
}
