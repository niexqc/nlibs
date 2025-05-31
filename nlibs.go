package nlibs

import (
	"reflect"

	"github.com/niexqc/nlibs/ntools"
	"github.com/robfig/cron/v3"
)

var FileDirExt = ntools.GetFileDirExt()

var HttpClientExt = ntools.GetHttpClientExt()

// 判断对象是否是数组或切片
func IsArrayOrSlice(v any) bool {
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// 返回一个支持至 秒 级别的 cron
func NewCronWithSeconds() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}

// 基础类型切片展开为为any切片
func Arr2ArrAny[T any](args []T) []any {
	anyArgs := make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}
	return anyArgs
}
