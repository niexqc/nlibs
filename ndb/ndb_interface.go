package ndb

import "github.com/niexqc/nlibs/ndb/sqlext"

type INdbWrapper interface {
	// 查询并生成动态对象返回
	//  dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1")
	//  val, err := sqlext.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
	SelectDyObj(sqlStr string, args ...any) (objVal *sqlext.NdbDyObj, err error)
	SelectDyObjList(sqlStr string, args ...any) (objValList []*sqlext.NdbDyObj, err error)
	// 查询单个字段单个值
	//  sqlStr:=select id from table where id=?
	//  str:=ndb.SelectOne[string](ndbw,sql,id)
	SelectOne(dest any, sqlStr string, args ...any) error
	SelectObj(dest any, sqlStr string, args ...any) error
	SelectList(dest any, sqlStr string, args ...any) error
	Exec(sqlStr string, args ...any) (rowsAffected int64, err error)
	Insert(sqlStr string, args ...any) (lastInsertId int64, err error)
}
