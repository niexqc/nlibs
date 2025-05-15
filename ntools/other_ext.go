package ntools

import (
	"fmt"
	"runtime"
)

// 匿名函数
func If3[T any](cond bool, ok, no T) T {
	if cond {
		return ok
	} else {
		return no
	}
}

// 判断OS是否是windows
func OsIsWindows() bool {
	return runtime.GOOS == "windows"
}

func FileSize2Str(byteSize int64) string {
	if byteSize > 1024*1024*1024 {
		return fmt.Sprintf("%.2fGb", float64(byteSize)/float64(1024*1024*1024))
	} else if byteSize > 1024*1024 {
		return fmt.Sprintf("%.2fMb", float64(byteSize)/float64(1024*1024))
	} else if byteSize > 1024 {
		return fmt.Sprintf("%.2fKb", float64(byteSize)/float64(1024))
	} else {
		return fmt.Sprintf("%dB", byteSize)
	}
}
