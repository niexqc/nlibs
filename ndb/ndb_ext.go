package ndb

import (
	"reflect"

	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
)

func NewNMysqlWrapper(conf *nyaml.YamlConfDb) *nmysql.NMysqlWrapper {
	return nmysql.NewNMysqlWrapper(conf)
}

func SelectOne[T any](ndbw INdbWrapper, sqlStr string, args ...any) (t *T, err error) {
	if mysqlNdb, ok := ndbw.(*nmysql.NMysqlWrapper); ok {
		obj := new(T)
		err = mysqlNdb.SelectOne(obj, sqlStr, args...)
		return obj, err
	} else {
		panic(nerror.NewRunTimeError(reflect.TypeOf(ndbw).Name() + "未实现 SelectOne"))
	}
}

func SelectObj[T any](ndbw INdbWrapper, sqlStr string, args ...any) (t *T, err error) {
	if mysqlNdb, ok := ndbw.(*nmysql.NMysqlWrapper); ok {
		obj := new(T)
		err = mysqlNdb.SelectObj(obj, sqlStr, args...)
		return obj, err
	} else {
		panic(nerror.NewRunTimeError(reflect.TypeOf(ndbw).Name() + "未实现 SelectObj"))
	}
}

func SelectList[T any](ndbw INdbWrapper, sqlStr string, args ...any) (tlist []*T, err error) {
	if mysqlNdb, ok := ndbw.(*nmysql.NMysqlWrapper); ok {
		objs := new([]*T)
		err = mysqlNdb.SelectList(objs, sqlStr, args...)
		return *objs, err
	} else {
		panic(nerror.NewRunTimeError(reflect.TypeOf(ndbw).Name() + "未实现 SelectList"))
	}
}
