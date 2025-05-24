package ntools

import (
	"fmt"
	"reflect"
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

// 判断any是否是指针，如果是指针-解引用
func AnyElem(val any, depth ...int) any {
	currentDepth := 0
	if len(depth) > 0 {
		currentDepth = depth[0]
		if currentDepth > 100 { // 防止无限递归
			return val
		}
	}
	if val == nil {
		return val
	}
	// 处理 reflect.Value 类型
	if reflect.TypeOf(val) == reflect.TypeOf(reflect.Value{}) {
		rv := val.(reflect.Value)
		if rv.CanInterface() {
			return AnyElem(rv.Interface(), currentDepth+1)
		}
		return rv // 不可导出时返回原始 reflect.Value
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return AnyElem(v.Elem().Interface(), currentDepth+1)
	}
	return v.Interface()
}
