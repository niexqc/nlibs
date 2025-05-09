package ndb

type INdbWrapper interface {
	SelectDyObj(sqlStr string, args ...any) (objVal any, err error)
	SelectDyObjList(sqlStr string, args ...any) (objValList []any, err error)

	SelectOne(dest any, sqlStr string, args ...any) error
	SelectObj(dest any, sqlStr string, args ...any) error
	SelectList(dest any, sqlStr string, args ...any) error
	Exec(sqlStr string, args ...any) (rowsAffected int64, err error)
	Insert(sqlStr string, args ...any) (lastInsertId int64, err error)
}
