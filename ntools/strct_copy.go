package ntools

import (
	"github.com/jinzhu/copier"
)

func StructCopy(src, dest any) error {
	return copier.Copy(dest, src)
}

func StructCopy2New[T any](src any) (t *T, err error) {
	nt := new(T)
	err = copier.Copy(nt, src)
	return nt, err
}

func StructCopy2NewPanicErr[T any](src any) *T {
	nt, err := StructCopy2New[T](src)
	if nil != err {
		panic(err)
	}
	return nt
}
