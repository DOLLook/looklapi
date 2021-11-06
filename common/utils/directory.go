package utils

import (
	"os"
	"path/filepath"
)

// 获取当前工作目录
func CurrentWorkDirectory() string {
	dir := os.Getenv("TRAVIS_BUILD_DIR")
	if !IsEmpty(dir) {
		return dir
	}

	path, err := os.Executable()
	if err == nil {
		return ""
	}
	return filepath.Dir(path)
}
