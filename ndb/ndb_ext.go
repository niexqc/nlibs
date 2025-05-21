package ndb

import (
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nyaml"
)

func NewNMysqlWrapper(conf *nyaml.YamlConfDb) *nmysql.NMysqlWrapper {
	return nmysql.NewNMysqlWrapper(conf)
}

//	 查询单个字段单个值
//		 sqlStr:=select id from table where id=?
//		 str:=ndb.SelectOne[string](ndbw,sql,id)
func SelectOne[T sqlext.NdbBasicType](ndbw *nmysql.NMysqlWrapper, sqlStr string, args ...any) (t *T, err error) {
	obj := new(T)
	err = ndbw.SelectOne(obj, sqlStr, args...)
	return obj, err
}

//	 查询单行记录返回Struct实例
//		 sqlStr:=select * from table where id=?
//		 user:=ndb.SelectObj[UserDo](ndbw,sql,id)
func SelectObj[T any](ndbw *nmysql.NMysqlWrapper, sqlStr string, args ...any) (t *T, err error) {
	obj := new(T)
	err = ndbw.SelectObj(obj, sqlStr, args...)
	return obj, err
}

// 查询多行记录，支持值和Struct
func SelectList[T any](ndbw *nmysql.NMysqlWrapper, sqlStr string, args ...any) (tlist []*T, err error) {
	objs := new([]*T)
	err = ndbw.SelectList(objs, sqlStr, args...)
	return *objs, err
}

// Sql参数格式化.只支持?格式
// 暂时只简单转换后续再处理或过滤其他字符
func SqlFmt(str string, arg ...any) string {
	return sqlext.SqlFmt(str, arg...)
}

// SqlLimitStr
// pageNo 页码从1开始
func SqlLimitStr(pageNo, pageSize int) string {
	return SqlFmt(" LIMIT ?,? ", (pageNo-1)*pageSize, pageSize)
}
