package ntools

import (
	"github.com/jinzhu/copier"
)

// 字段类型不匹配时，赋0值
func StructCopy(src, dest any) error {
	return copier.Copy(dest, src)
}

// 字段类型不匹配时，赋0值
func StructCopy2New[T any](src any) (t *T, err error) {
	nt := new(T)
	err = copier.Copy(nt, src)
	return nt, err
}

// 字段类型不匹配时，赋0值
func StructCopy2NewPanicErr[T any](src any) *T {
	nt, err := StructCopy2New[T](src)
	if nil != err {
		panic(err)
	}
	return nt
}
