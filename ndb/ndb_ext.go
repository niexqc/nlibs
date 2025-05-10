package ndb

import (
	"reflect"

	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
)

func NewNMysqlWrapper(conf *nyaml.YamlConfDb) *nmysql.NMysqlWrapper {
	return nmysql.NewNMysqlWrapper(conf)
}

func SelectOne[T sqlext.NdbBasicType](ndbw INdbWrapper, sqlStr string, args ...any) (t *T, err error) {
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

// Sql参数格式化.只支持?格式
// 暂时只简单转换后续再处理或过滤其他字符
func SqlFmt(str string, arg ...any) string {
	return sqlext.SqlFmt(str, arg...)
}
