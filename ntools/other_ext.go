package ntools

import "runtime"

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
